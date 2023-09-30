package aicra_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra"
	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/aicra/runtime"
)

func printEscaped(raw string) string {
	raw = strings.ReplaceAll(raw, "\n", "\\n")
	raw = strings.ReplaceAll(raw, "\r", "\\r")
	return raw
}

func bind(method, path string, fn http.HandlerFunc) func(*aicra.Builder) error {
	return func(b *aicra.Builder) error {
		return b.Bind(method, path, fn)
	}
}

func TestHandlerWith(t *testing.T) {
	builder := &aicra.Builder{}

	// build @n middlewares that take data from context and increment it
	n := 1024

	type ckey int
	const key ckey = 0

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// first time -> store 1
			value := r.Context().Value(key)
			if value == nil {
				r = r.WithContext(context.WithValue(r.Context(), key, int(1)))
				next.ServeHTTP(w, r)
				return
			}

			// get value and increment
			cast, ok := value.(int)
			require.True(t, ok, "cannot cast context data to int")
			cast++
			r = r.WithContext(context.WithValue(r.Context(), key, cast))
			next.ServeHTTP(w, r)
		})
	}

	// add middleware @n times
	for i := 0; i < n; i++ {
		builder.With(middleware)
	}

	cnf := strings.NewReader(`[ { "method": "GET", "path": "/path", "scope": [[]], "info": "info", "in": {}, "out": {} } ]`)
	err := builder.Setup(cnf)
	require.NotNil(t, err)

	pathHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// write value from middlewares into response
		value := r.Context().Value(key)
		if value == nil {
			t.Fatalf("nothing found in context")
		}
		cast, ok := value.(int)
		if !ok {
			t.Fatalf("cannot cast context data to int")
		}
		// write to response
		w.Write([]byte(fmt.Sprintf("#%d#", cast)))
	})

	err = builder.Bind(http.MethodGet, "/path", pathHandler)
	require.NoError(t, err)

	handler, err := builder.Build(config.Validators{})
	require.NoError(t, err)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/path", &bytes.Buffer{})

	// test request
	handler.ServeHTTP(response, request)
	require.NotNil(t, response.Body, "body")
	token := fmt.Sprintf("#%d#", n)
	require.Containsf(t, response.Body.String(), token, "expected %q to be in response <%s>", token, response.Body.String())

}

func TestHandlerWithAuth(t *testing.T) {

	tt := []struct {
		name        string
		config      string
		permissions []string
		granted     bool
	}{
		{
			name:        "provide only requirement A",
			config:      `[ { "method": "GET", "path": "/path", "scope": [["A"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A"},
			granted:     true,
		},
		{
			name:        "missing requirement",
			config:      `[ { "method": "GET", "path": "/path", "scope": [["A"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{},
			granted:     false,
		},
		{
			name:        "missing requirements",
			config:      `[ { "method": "GET", "path": "/path", "scope": [["A", "B"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{},
			granted:     false,
		},
		{
			name:        "missing some requirements",
			config:      `[ { "method": "GET", "path": "/path", "scope": [["A", "B"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A"},
			granted:     false,
		},
		{
			name:        "provide requirements",
			config:      `[ { "method": "GET", "path": "/path", "scope": [["A", "B"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A", "B"},
			granted:     true,
		},
		{
			name:        "missing OR requirements",
			config:      `[ { "method": "GET", "path": "/path", "scope": [["A"], ["B"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"C"},
			granted:     false,
		},
		{
			name:        "provide 1 OR requirement",
			config:      `[ { "method": "GET", "path": "/path", "scope": [["A"], ["B"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A"},
			granted:     true,
		},
		{
			name:        "provide both OR requirements",
			config:      `[ { "method": "GET", "path": "/path", "scope": [["A"], ["B"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A", "B"},
			granted:     true,
		},
		{
			name:        "missing composite OR requirements",
			config:      `[ { "method": "GET", "path": "/path", "scope": [["A", "B"], ["C", "D"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{},
			granted:     false,
		},
		{
			name:        "missing partial composite OR requirements",
			config:      `[ { "method": "GET", "path": "/path", "scope": [["A", "B"], ["C", "D"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A", "C"},
			granted:     false,
		},
		{
			name:        "provide 1 composite OR requirement",
			config:      `[ { "method": "GET", "path": "/path", "scope": [["A", "B"], ["C", "D"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A", "B", "C"},
			granted:     true,
		},
		{
			name:        "provide both composite OR requirements",
			config:      `[ { "method": "GET", "path": "/path", "scope": [["A", "B"], ["C", "D"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A", "B", "C", "D"},
			granted:     true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			builder := &aicra.Builder{}

			// tester middleware (last executed)
			builder.WithContext(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := api.Extract(r.Context()).Auth
					require.NotNil(t, a, "cannot access api.Auth form request context")

					require.Equal(t, tc.granted, a.Granted(), "auth granted")
					next.ServeHTTP(w, r)
				})
			})

			builder.WithContext(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := api.Extract(r.Context()).Auth
					require.NotNil(t, a, "cannot access api.Auth form request context")

					a.Active = tc.permissions
					next.ServeHTTP(w, r)
				})
			})

			err := builder.Setup(strings.NewReader(tc.config))
			require.NoError(t, err)

			pathHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				runtime.Respond(w, nil, api.ErrNotImplemented)
			})

			err = builder.Bind(http.MethodGet, "/path", pathHandler)
			require.NoError(t, err)

			handler, err := builder.Build(config.Validators{})
			require.NoError(t, err)

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/path", &bytes.Buffer{})

			// test request
			handler.ServeHTTP(response, request)
			require.NotNil(t, response.Body, "body")

		})
	}

}

func TestHandlerPermissionError(t *testing.T) {
	noOpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runtime.Respond(w, nil, api.ErrNotImplemented)
	})

	tt := []struct {
		name        string
		uri, body   string
		config      string
		binder      func(*aicra.Builder) error
		permissions []string
		err         api.Err
	}{
		{
			name:        "permission fulfilled",
			uri:         "/path",
			config:      `[ { "method": "GET", "path": "/path", "scope": [["A"]], "info": "info", "in": {}, "out": {} } ]`,
			binder:      bind("GET", "/path", noOpHandler),
			permissions: []string{"A"},
			err:         api.ErrNotImplemented,
		},
		{
			name:        "missing permission",
			uri:         "/path",
			config:      `[ { "method": "GET", "path": "/path", "scope": [["A"]], "info": "info", "in": {}, "out": {} } ]`,
			binder:      bind("GET", "/path", noOpHandler),
			permissions: []string{},
			err:         api.ErrForbidden,
		},
		// check that permission errors are raised:
		// AFTER uri params
		// BEFORE query and body params
		{
			name: "permission with wrong uri param",
			uri:  "/path/abc",
			config: `[ {
				"method": "GET",
				"path": "/path/{uid}",
				"scope": [["missing"]],
				"info": "info",
				"in": {
					"{uid}": { "info": "user id", "type": "uint", "name": "UserID" }
				},
				"out": {}
			} ]`,
			binder:      bind("GET", "/path/{uid}", noOpHandler),
			permissions: []string{},
			err:         api.ErrUnknownService,
		},
		{
			name: "permission with wrong query param",
			uri:  "/path?uid=invalid-type",
			config: `[ {
				"method": "GET",
				"path": "/path",
				"scope": [["missing"]],
				"info": "info",
				"in": {
					"GET@uid": { "info": "user id", "type": "uint", "name": "UserID" }
				},
				"out": {}
			} ]`,
			binder:      bind("GET", "/path", noOpHandler),
			permissions: []string{},
			err:         api.ErrForbidden,
		},
		{
			name: "permission with wrong body param",
			uri:  "/path",
			body: "uid=invalid-type",
			config: `[ {
				"method": "GET",
				"path": "/path",
				"scope": [["missing"]],
				"info": "info",
				"in": {
					"uid": { "info": "user id", "type": "uint", "name": "UserID" }
				},
				"out": {}
			} ]`,
			binder:      bind("GET", "/path", noOpHandler),
			permissions: []string{},
			err:         api.ErrForbidden,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			builder := &aicra.Builder{}

			// add active permissions
			builder.WithContext(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := api.Extract(r.Context()).Auth
					require.NotNil(t, a, "cannot access api.Auth form request context")

					a.Active = tc.permissions
					next.ServeHTTP(w, r)
				})
			})

			err := builder.Setup(strings.NewReader(tc.config))
			require.NoError(t, err)

			err = tc.binder(builder)
			require.NoError(t, err)

			handler, err := builder.Build(config.Validators{})
			require.NoError(t, err)

			var (
				body     = strings.NewReader(tc.body)
				response = httptest.NewRecorder()
				request  = httptest.NewRequest(http.MethodGet, tc.uri, body)
			)

			// test request
			handler.ServeHTTP(response, request)
			require.NotNil(t, response.Body, "body")

			expectedStatus := api.GetErrorStatus(tc.err)
			require.Equalf(t, expectedStatus, response.Result().StatusCode, "http status")
		})
	}

}

func TestHandlerDynamicScope(t *testing.T) {
	noOpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runtime.Respond(w, nil, api.ErrNotImplemented)
	})

	tt := []struct {
		name        string
		config      string
		binder      func(*aicra.Builder) error
		url         string
		body        string
		permissions []string
		granted     bool
	}{
		{
			name: "replace one granted",
			config: `[
				{
					"method": "POST",
					"path": "/path/{id}",
					"info": "info",
					"scope": [["user[Input1]"]],
					"in": {
						"{id}": { "info": "info", "name": "Input1", "type": "uint" }
					},
					"out": {}
				}
			]`,
			binder:      bind("POST", "/path/{id}", noOpHandler),
			url:         "/path/123",
			body:        ``,
			permissions: []string{"user[123]"},
			granted:     true,
		},
		{
			name: "replace one mismatch",
			config: `[
				{
					"method": "POST",
					"path": "/path/{id}",
					"info": "info",
					"scope": [["user[Input1]"]],
					"in": {
						"{id}": { "info": "info", "name": "Input1", "type": "uint" }
					},
					"out": {}
				}
			]`,
			binder:      bind("POST", "/path/{id}", noOpHandler),
			url:         "/path/666",
			body:        ``,
			permissions: []string{"user[123]"},
			granted:     false,
		},
		{
			name: "replace one valid dot separated",
			config: `[
				{
					"method": "POST",
					"path": "/path/{id}",
					"info": "info",
					"scope": [["prefix.user[User].suffix"]],
					"in": {
						"{id}": { "info": "info", "name": "User", "type": "uint" }
					},
					"out": {}
				}
			]`,
			binder:      bind("POST", "/path/{id}", noOpHandler),
			url:         "/path/123",
			body:        ``,
			permissions: []string{"prefix.user[123].suffix"},
			granted:     true,
		},
		{
			name: "replace two valid dot separated",
			config: `[
				{
					"method": "POST",
					"path": "/prefix/{pid}/user/{uid}",
					"info": "info",
					"scope": [["prefix[Prefix].user[User].suffix"]],
					"in": {
						"{pid}": { "info": "info", "name": "Prefix", "type": "uint" },
						"{uid}": { "info": "info", "name": "User", "type": "uint" }
					},
					"out": {}
				}
			]`, binder: bind("POST", "/prefix/{pid}/user/{uid}", noOpHandler),
			url:         "/prefix/123/user/456",
			body:        ``,
			permissions: []string{"prefix[123].user[456].suffix"},
			granted:     true,
		},
		{
			name: "replace two invalid dot separated",
			config: `[
				{
					"method": "POST",
					"path": "/prefix/{pid}/user/{uid}",
					"info": "info",
					"scope": [["prefix[Prefix].user[User].suffix"]],
					"in": {
						"{pid}": { "info": "info", "name": "Prefix", "type": "uint" },
						"{uid}": { "info": "info", "name": "User", "type": "uint" }
					},
					"out": {}
				}
			]`,
			binder:      bind("POST", "/prefix/{pid}/user/{uid}", noOpHandler),
			url:         "/prefix/123/user/666",
			body:        ``,
			permissions: []string{"prefix[123].user[456].suffix"},
			granted:     false,
		},
		{
			name: "replace three valid dot separated",
			config: `[
				{
					"method": "POST",
					"path": "/prefix/{pid}/user/{uid}/suffix/{sid}",
					"info": "info",
					"scope": [["prefix[Prefix].user[User].suffix[Suffix]"]],
					"in": {
						"{pid}": { "info": "info", "name": "Prefix", "type": "uint" },
						"{uid}": { "info": "info", "name": "User", "type": "uint" },
						"{sid}": { "info": "info", "name": "Suffix", "type": "uint" }
					},
					"out": {}
				}
			]`,
			binder:      bind("POST", "/prefix/{pid}/user/{uid}/suffix/{sid}", noOpHandler),
			url:         "/prefix/123/user/456/suffix/789",
			body:        ``,
			permissions: []string{"prefix[123].user[456].suffix[789]"},
			granted:     true,
		},
		{
			name: "replace three invalid dot separated",
			config: `[
				{
					"method": "POST",
					"path": "/prefix/{pid}/user/{uid}/suffix/{sid}",
					"info": "info",
					"scope": [["prefix[Prefix].user[User].suffix[Suffix]"]],
					"in": {
						"{pid}": { "info": "info", "name": "Prefix", "type": "uint" },
						"{uid}": { "info": "info", "name": "User", "type": "uint" },
						"{sid}": { "info": "info", "name": "Suffix", "type": "uint" }
					},
					"out": {}
				}
			]`,
			binder:      bind("POST", "/prefix/{pid}/user/{uid}/suffix/{sid}", noOpHandler),
			url:         "/prefix/123/user/666/suffix/789",
			body:        ``,
			permissions: []string{"prefix[123].user[456].suffix[789]"},
			granted:     false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			builder := &aicra.Builder{}

			// tester middleware (last executed)
			builder.WithContext(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := api.Extract(r.Context()).Auth
					require.NotNil(t, a, "cannot access api.Auth form request context")
					require.Equal(t, tc.granted, a.Granted(), "auth granted")
					next.ServeHTTP(w, r)
				})
			})

			// update permissions
			builder.WithContext(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := api.Extract(r.Context()).Auth
					require.NotNil(t, a, "cannot access api.Auth form request context")
					a.Active = tc.permissions
					next.ServeHTTP(w, r)
				})
			})

			err := builder.Setup(strings.NewReader(tc.config))
			require.NoError(t, err)

			err = tc.binder(builder)
			require.NoError(t, err)

			handler, err := builder.Build(config.Validators{})
			require.NoError(t, err)

			response := httptest.NewRecorder()
			body := strings.NewReader(tc.body)
			request := httptest.NewRequest(http.MethodPost, tc.url, body)

			// test request
			handler.ServeHTTP(response, request)
			require.NotNil(t, response.Body, "body")
		})
	}

}

func TestHandlerServiceErrors(t *testing.T) {
	noOpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runtime.Respond(w, nil, api.ErrNotImplemented)
	})

	tt := []struct {
		name   string
		config string
		// handler
		binder func(*aicra.Builder) error
		// request
		method, url string
		contentType string
		body        string
		permissions []string
		err         error
		errReason   string
	}{
		// service match
		{
			name: "unknown service method",
			config: `[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {},
					"out": {}
				}
			]`,
			binder:      bind("GET", "/", noOpHandler),
			method:      http.MethodPost,
			url:         "/",
			body:        ``,
			permissions: []string{},
			err:         api.ErrUnknownService,
		},
		{
			name: "unknown service path",
			config: `[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {},
					"out": {}
				}
			]`,
			binder:      bind("GET", "/", noOpHandler),
			method:      http.MethodGet,
			url:         "/invalid",
			body:        ``,
			permissions: []string{},
			err:         api.ErrUnknownService,
		},
		{
			name: "valid empty service",
			config: `[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {},
					"out": {}
				}
			]`,
			binder:      bind("GET", "/", noOpHandler),
			method:      http.MethodGet,
			url:         "/",
			body:        ``,
			permissions: []string{},
			err:         nil,
		},

		// invalid uri param -> unknown service
		{
			name: "invalid uri param",
			config: `[
				{
					"method": "GET",
					"path": "/a/{id}/b",
					"info": "info",
					"scope": [],
					"in":  {
						"{id}": { "info": "info", "type": "int", "name": "ID" }
					},
					"out": {}
				}
			]`,
			binder:      bind("GET", "/a/{id}/b", noOpHandler),
			method:      http.MethodGet,
			url:         "/a/invalid/b",
			body:        ``,
			permissions: []string{},
			err:         api.ErrUnknownService,
		},

		// query param
		{
			name: "missing query param",
			config: `[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {
						"GET@id": { "info": "info", "type": "int", "name": "ID" }
					},
					"out": {}
				}
			]`,
			binder:      bind("GET", "/", noOpHandler),
			method:      http.MethodGet,
			url:         "/",
			body:        ``,
			permissions: []string{},
			err:         api.ErrMissingParam,
			errReason:   fmt.Sprintf("ID: %s", api.ErrMissingParam.Error()),
		},
		{
			name: "invalid query param",
			config: `[
				{
					"method": "GET",
					"path": "/a",
					"info": "info",
					"scope": [],
					"in":  {
						"GET@id": { "info": "info", "type": "int", "name": "ID" }
					},
					"out": {}
				}
			]`,
			binder:      bind("GET", "/a", noOpHandler),
			method:      http.MethodGet,
			url:         "/a?id=abc",
			body:        ``,
			permissions: []string{},
			err:         api.ErrInvalidParam,
			errReason:   fmt.Sprintf("ID: %s", api.ErrInvalidParam.Error()),
		},
		{
			name: "query unexpected slice param",
			config: `[
				{
					"method": "GET",
					"path": "/a",
					"info": "info",
					"scope": [],
					"in":  {
						"GET@id": { "info": "info", "type": "int", "name": "ID" }
					},
					"out": {}
				}
			]`,
			binder:      bind("GET", "/a", noOpHandler),
			method:      http.MethodGet,
			url:         "/a?id=123&id=456",
			body:        ``,
			permissions: []string{},
			err:         api.ErrInvalidParam,
			errReason:   fmt.Sprintf("ID: %s", api.ErrInvalidParam.Error()),
		},
		{
			name: "valid query param",
			config: `[
				{
					"method": "GET",
					"path": "/a",
					"info": "info",
					"scope": [],
					"in":  {
						"GET@id": { "info": "info", "type": "int", "name": "ID" }
					},
					"out": {}
				}
			]`,
			binder:      bind("GET", "/a", noOpHandler),
			method:      http.MethodGet,
			url:         "/a?id=123",
			body:        ``,
			permissions: []string{},
			err:         nil,
		},

		// json param
		{
			name: "missing json param",
			config: `[
				{
					"method": "POST",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {
						"id": { "info": "info", "type": "int", "name": "ID" }
					},
					"out": {}
				}
			]`,
			binder:      bind("POST", "/", noOpHandler),
			contentType: "application/json",
			method:      http.MethodPost,
			url:         "/",
			body:        ``,
			permissions: []string{},
			err:         api.ErrMissingParam,
			errReason:   fmt.Sprintf("ID: %s", api.ErrMissingParam.Error()),
		},
		{
			name: "invalid json param",
			config: `[
				{
					"method": "POST",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {
						"id": { "info": "info", "type": "int", "name": "ID" }
					},
					"out": {}
				}
			]`,
			binder:      bind("POST", "/", noOpHandler),
			contentType: "application/json",
			method:      http.MethodPost,
			url:         "/",
			body:        `{ "id": "invalid type" }`,
			permissions: []string{},
			err:         api.ErrInvalidParam,
			errReason:   fmt.Sprintf("ID: %s", api.ErrInvalidParam.Error()),
		},
		{
			name: "valid json param",
			config: `[
				{
					"method": "POST",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {
						"id": { "info": "info", "type": "int", "name": "ID" }
					},
					"out": {}
				}
			]`,
			binder:      bind("POST", "/", noOpHandler),
			contentType: "application/json",
			method:      http.MethodPost,
			url:         "/",
			body:        `{ "id": 123 }`,
			permissions: []string{},
			err:         nil,
		},

		// urlencoded param
		{
			name: "missing urlencoded param",
			config: `[
				{
					"method": "POST",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {
						"id": { "info": "info", "type": "int", "name": "ID" }
					},
					"out": {}
				}
			]`,
			binder:      bind("POST", "/", noOpHandler),
			contentType: "application/x-www-form-urlencoded",
			method:      http.MethodPost,
			url:         "/",
			body:        ``,
			permissions: []string{},
			err:         api.ErrMissingParam,
			errReason:   fmt.Sprintf("ID: %s", api.ErrMissingParam.Error()),
		},
		{
			name: "invalid urlencoded param",
			config: `[
				{
					"method": "POST",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {
						"id": { "info": "info", "type": "int", "name": "ID" }
					},
					"out": {}
				}
			]`,
			binder:      bind("POST", "/", noOpHandler),
			contentType: "application/x-www-form-urlencoded",
			method:      http.MethodPost,
			url:         "/",
			body:        `id=abc`,
			permissions: []string{},
			err:         api.ErrInvalidParam,
			errReason:   fmt.Sprintf("ID: %s", api.ErrInvalidParam.Error()),
		},
		{
			name: "valid urlencoded param",
			config: `[
				{
					"method": "POST",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {
						"id": { "info": "info", "type": "int", "name": "ID" }
					},
					"out": {}
				}
			]`,
			binder:      bind("POST", "/", noOpHandler),
			contentType: "application/x-www-form-urlencoded",
			method:      http.MethodPost,
			url:         "/",
			body:        `id=123`,
			permissions: []string{},
			err:         nil,
		},

		// formdata param
		{
			name: "missing multipart param",
			config: `[
				{
					"method": "POST",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {
						"id": { "info": "info", "type": "int", "name": "ID" }
					},
					"out": {}
				}
			]`,
			binder:      bind("POST", "/", noOpHandler),
			contentType: "multipart/form-data; boundary=xxx",
			method:      http.MethodPost,
			url:         "/",
			body:        ``,
			permissions: []string{},
			err:         api.ErrMissingParam,
			errReason:   fmt.Sprintf("ID: %s", api.ErrMissingParam.Error()),
		},
		{
			name: "invalid multipart param",
			config: `[
				{
					"method": "POST",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {
						"id": { "info": "info", "type": "int", "name": "ID" }
					},
					"out": {}
				}
			]`,
			binder:      bind("POST", "/", noOpHandler),
			contentType: "multipart/form-data; boundary=xxx",
			method:      http.MethodPost,
			url:         "/",
			body: `--xxx
Content-Disposition: form-data; name="id"

abc
--xxx--`,
			permissions: []string{},
			err:         api.ErrInvalidParam,
			errReason:   fmt.Sprintf("ID: %s", api.ErrInvalidParam.Error()),
		},
		{
			name: "valid multipart param",
			config: `[
				{
					"method": "POST",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {
						"id": { "info": "info", "type": "int", "name": "ID" }
					},
					"out": {}
				}
			]`,
			binder:      bind("POST", "/", noOpHandler),
			contentType: "multipart/form-data; boundary=xxx",
			method:      http.MethodPost,
			url:         "/",
			body: `--xxx
Content-Disposition: form-data; name="id"

123
--xxx--`,
			permissions: []string{},
			err:         nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			builder := &aicra.Builder{}
			err := builder.Setup(strings.NewReader(tc.config))
			require.NoError(t, err)

			err = tc.binder(builder)
			require.NoError(t, err)

			handler, err := builder.Build(config.Validators{})
			require.NoError(t, err)

			var (
				response = httptest.NewRecorder()
				body     = strings.NewReader(tc.body)
				request  = httptest.NewRequest(tc.method, tc.url, body)
			)
			if len(tc.contentType) > 0 {
				request.Header.Add("Content-Type", tc.contentType)
			}

			// test request
			handler.ServeHTTP(response, request)
			require.Equalf(t, api.GetErrorStatus(tc.err), response.Result().StatusCode, "http status")

			if len(tc.errReason) < 1 {
				return
			}

			type JSONError struct {
				Status string `json:"status"`
			}
			var parsedError JSONError
			err = json.NewDecoder(response.Body).Decode(&parsedError)
			require.NoError(t, err, "parse body")
			require.Equal(t, tc.errReason, parsedError.Status, "error reason")
		})
	}
}

func TestHandlerResponse(t *testing.T) {
	tt := []struct {
		name   string
		config string

		// handler
		binder func(*aicra.Builder) error
		// request
		method, uri, body string

		response string
	}{
		{
			name: "nil error",
			config: `[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {},
					"out": {}
				}
			]`,
			binder: bind("GET", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				runtime.Respond(w, nil, nil)
			})),
			method:   http.MethodGet,
			uri:      "/",
			response: `{"status":"all right"}`,
		},
		{
			name: "non-nil error",
			config: `[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {},
					"out": {}
				}
			]`,
			binder: bind("GET", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				runtime.Respond(w, nil, api.ErrNotImplemented)
			})),
			method:   http.MethodGet,
			uri:      "/",
			response: `{"status":"not implemented"}`,
		},

		{
			name: "nil int output",
			config: `[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {},
					"out": {
						"id": { "name": "ID", "info": "info", "type": "uint" }
					}
				}
			]`,
			binder: bind("GET", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				runtime.Respond(w, nil, nil)
			})),
			method:   http.MethodGet,
			uri:      "/",
			response: `{"status":"all right"}`,
		},
		{
			name: "non-nil int output",
			config: `[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {},
					"out": {
						"id": { "name": "ID", "info": "info", "type": "uint" }
					}
				}
			]`,
			binder: bind("GET", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				runtime.Respond(w, map[string]any{"id": 123}, nil)
			})),
			method:   http.MethodGet,
			uri:      "/",
			response: `{"id":123,"status":"all right"}`,
		},
		{
			name: "nil int outputs",
			config: `[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {},
					"out": {
						"a": { "name": "A", "info": "info", "type": "uint" },
						"z": { "name": "Z", "info": "info", "type": "uint" }
					}
				}
			]`,
			binder: bind("GET", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				runtime.Respond(w, nil, api.ErrForbidden)
			})),
			method:   http.MethodGet,
			uri:      "/",
			response: `{"status":"forbidden"}`,
		},
		{
			name: "int outputs surrounding error",
			config: `[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {},
					"out": {
						"a": { "name": "A", "info": "info", "type": "uint" },
						"z": { "name": "Z", "info": "info", "type": "uint" }
					}
				}
			]`,
			binder: bind("GET", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data := map[string]any{
					"a": 123,
					"z": 456,
				}
				runtime.Respond(w, data, nil)
			})),
			method:   http.MethodGet,
			uri:      "/",
			response: `{"a":123,"status":"all right","z":456}`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			builder := &aicra.Builder{}

			err := builder.Setup(strings.NewReader(tc.config))
			require.NoError(t, err)

			err = tc.binder(builder)
			require.NoError(t, err)

			handler, err := builder.Build(config.Validators{})
			require.NoError(t, err)

			var (
				response = httptest.NewRecorder()
				body     = strings.NewReader(tc.body)
				request  = httptest.NewRequest(tc.method, tc.uri, body)
			)
			request.Header.Add("Content-Type", "application/json")

			// test request
			handler.ServeHTTP(response, request)
			require.NotNil(t, response.Body, "body")
			require.Equalf(t, tc.response, response.Body.String(), "response")
		})
	}
}

func TestHandlerRequestTooLarge(t *testing.T) {
	tt := []struct {
		name              string
		uriMax, uriSize   int
		bodyMax, bodySize int
		err               error
	}{
		{
			name:     "defaults -1",
			uriSize:  aicra.DefaultURILimit - 1,
			bodySize: aicra.DefaultBodyLimit - 1,
			err:      api.ErrUnknownService,
		},
		{
			name:     "defaults eq",
			uriSize:  aicra.DefaultURILimit,
			bodySize: aicra.DefaultBodyLimit,
			err:      api.ErrUnknownService,
		},
		{
			name:    "defaults uri",
			uriSize: aicra.DefaultURILimit + 1,
			err:     api.ErrURITooLong,
		},
		{
			name:     "defaults body",
			bodySize: aicra.DefaultBodyLimit + 1,
			err:      api.ErrBodyTooLarge,
		},
		{
			name:     "defaults both",
			uriSize:  aicra.DefaultURILimit + 1,
			bodySize: aicra.DefaultBodyLimit + 1,
			err:      api.ErrURITooLong,
		},

		{
			name:    "unlimited uri",
			uriMax:  -1,
			uriSize: aicra.DefaultURILimit + 1,
			err:     api.ErrUnknownService,
		},
		{
			name:     "unlimited body",
			bodyMax:  -1,
			bodySize: aicra.DefaultBodyLimit + 1,
			err:      api.ErrUnknownService,
		},
		{
			name:    "custom uri ok",
			uriMax:  50,
			uriSize: 50,
			err:     api.ErrUnknownService,
		},
		{
			name:    "custom uri",
			uriMax:  50,
			uriSize: 51,
			err:     api.ErrURITooLong,
		},
		{
			name:     "custom body ok",
			bodyMax:  50,
			bodySize: 50,
			err:      api.ErrUnknownService,
		},
		{
			name:     "custom body",
			bodyMax:  50,
			bodySize: 51,
			err:      api.ErrBodyTooLarge,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			b := &aicra.Builder{}

			err := b.Setup(strings.NewReader(`[]`))
			require.NoError(t, err)

			b.SetURILimit(tc.uriMax)
			b.SetBodyLimit(int64(tc.bodyMax))

			handler, err := b.Build(config.Validators{})
			require.NoError(t, err)

			srv := httptest.NewServer(handler)
			defer srv.Close()

			host := fmt.Sprintf("%s/", srv.URL)

			// build fake uri and body according to test sizes
			var (
				fakeURI  = strings.Repeat("a", tc.uriSize)
				fakeBody = strings.Repeat("a", tc.bodySize)
			)
			// remove 1 to take the '/' into account
			if len(fakeURI) > 0 {
				fakeURI = strings.TrimSuffix(fakeURI, "a")
			}

			req, err := http.NewRequest(
				"POST",
				host+fakeURI,
				strings.NewReader(fakeBody),
			)
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			var expect int = http.StatusOK
			if tc.err != nil {
				cast, ok := tc.err.(api.Err)
				require.True(t, ok, "invalid error type")
				expect = cast.Status()
			}
			require.Equal(t, expect, res.StatusCode, "http status")
		})
	}
}
