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

	"github.com/xdrm-io/aicra"
	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/validator"
)

func printEscaped(raw string) string {
	raw = strings.ReplaceAll(raw, "\n", "\\n")
	raw = strings.ReplaceAll(raw, "\r", "\\r")
	return raw
}

func addDefaultTypes(b *aicra.Builder) error {
	if err := b.Validate(validator.AnyType{}); err != nil {
		return err
	}
	if err := b.Validate(validator.BoolType{}); err != nil {
		return err
	}
	if err := b.Validate(validator.FloatType{}); err != nil {
		return err
	}
	if err := b.Validate(validator.IntType{}); err != nil {
		return err
	}
	if err := b.Validate(validator.StringType{}); err != nil {
		return err
	}
	if err := b.Validate(validator.UintType{}); err != nil {
		return err
	}
	return nil
}

func TestHandler_With(t *testing.T) {
	builder := &aicra.Builder{}
	if err := addDefaultTypes(builder); err != nil {
		t.Fatalf("unexpected error <%v>", err)
	}

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
			if !ok {
				t.Fatalf("value is not an int")
			}
			cast++
			r = r.WithContext(context.WithValue(r.Context(), key, cast))
			next.ServeHTTP(w, r)
		})
	}

	// add middleware @n times
	for i := 0; i < n; i++ {
		builder.With(middleware)
	}

	config := strings.NewReader(`[ { "method": "GET", "path": "/path", "scope": [[]], "info": "info", "in": {}, "out": {} } ]`)
	err := builder.Setup(config)
	if err != nil {
		t.Fatalf("setup: unexpected error <%v>", err)
	}

	pathHandler := func(ctx context.Context) (*struct{}, api.Err) {
		// write value from middlewares into response
		value := ctx.Value(key)
		if value == nil {
			t.Fatalf("nothing found in context")
		}
		cast, ok := value.(int)
		if !ok {
			t.Fatalf("cannot cast context data to int")
		}
		// write to response
		api.GetResponseWriter(ctx).Write([]byte(fmt.Sprintf("#%d#", cast)))

		return nil, api.ErrSuccess
	}

	if err := builder.Bind(http.MethodGet, "/path", pathHandler); err != nil {
		t.Fatalf("bind: unexpected error <%v>", err)
	}

	handler, err := builder.Build()
	if err != nil {
		t.Fatalf("build: unexpected error <%v>", err)
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/path", &bytes.Buffer{})

	// test request
	handler.ServeHTTP(response, request)
	if response.Body == nil {
		t.Fatalf("response has no body")
	}
	token := fmt.Sprintf("#%d#", n)
	if !strings.Contains(response.Body.String(), token) {
		t.Fatalf("expected '%s' to be in response <%s>", token, response.Body.String())
	}

}

func TestHandler_WithAuth(t *testing.T) {

	tt := []struct {
		name        string
		manifest    string
		permissions []string
		granted     bool
	}{
		{
			name:        "provide only requirement A",
			manifest:    `[ { "method": "GET", "path": "/path", "scope": [["A"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A"},
			granted:     true,
		},
		{
			name:        "missing requirement",
			manifest:    `[ { "method": "GET", "path": "/path", "scope": [["A"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{},
			granted:     false,
		},
		{
			name:        "missing requirements",
			manifest:    `[ { "method": "GET", "path": "/path", "scope": [["A", "B"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{},
			granted:     false,
		},
		{
			name:        "missing some requirements",
			manifest:    `[ { "method": "GET", "path": "/path", "scope": [["A", "B"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A"},
			granted:     false,
		},
		{
			name:        "provide requirements",
			manifest:    `[ { "method": "GET", "path": "/path", "scope": [["A", "B"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A", "B"},
			granted:     true,
		},
		{
			name:        "missing OR requirements",
			manifest:    `[ { "method": "GET", "path": "/path", "scope": [["A"], ["B"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"C"},
			granted:     false,
		},
		{
			name:        "provide 1 OR requirement",
			manifest:    `[ { "method": "GET", "path": "/path", "scope": [["A"], ["B"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A"},
			granted:     true,
		},
		{
			name:        "provide both OR requirements",
			manifest:    `[ { "method": "GET", "path": "/path", "scope": [["A"], ["B"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A", "B"},
			granted:     true,
		},
		{
			name:        "missing composite OR requirements",
			manifest:    `[ { "method": "GET", "path": "/path", "scope": [["A", "B"], ["C", "D"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{},
			granted:     false,
		},
		{
			name:        "missing partial composite OR requirements",
			manifest:    `[ { "method": "GET", "path": "/path", "scope": [["A", "B"], ["C", "D"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A", "C"},
			granted:     false,
		},
		{
			name:        "provide 1 composite OR requirement",
			manifest:    `[ { "method": "GET", "path": "/path", "scope": [["A", "B"], ["C", "D"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A", "B", "C"},
			granted:     true,
		},
		{
			name:        "provide both composite OR requirements",
			manifest:    `[ { "method": "GET", "path": "/path", "scope": [["A", "B"], ["C", "D"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A", "B", "C", "D"},
			granted:     true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			builder := &aicra.Builder{}
			if err := addDefaultTypes(builder); err != nil {
				t.Fatalf("unexpected error <%v>", err)
			}

			// tester middleware (last executed)
			builder.WithContext(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := api.GetAuth(r.Context())
					if a == nil {
						t.Fatalf("cannot access api.Auth form request context")
					}

					if a.Granted() == tc.granted {
						return
					}
					if a.Granted() {
						t.Fatalf("unexpected granted auth")
					} else {
						t.Fatalf("expected granted auth")
					}
					next.ServeHTTP(w, r)
				})
			})

			builder.WithContext(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := api.GetAuth(r.Context())
					if a == nil {
						t.Fatalf("cannot access api.Auth form request context")
					}

					a.Active = tc.permissions
					next.ServeHTTP(w, r)
				})
			})

			err := builder.Setup(strings.NewReader(tc.manifest))
			if err != nil {
				t.Fatalf("setup: unexpected error <%v>", err)
			}

			pathHandler := func(ctx context.Context) (*struct{}, api.Err) {
				return nil, api.ErrNotImplemented
			}

			if err := builder.Bind(http.MethodGet, "/path", pathHandler); err != nil {
				t.Fatalf("bind: unexpected error <%v>", err)
			}

			handler, err := builder.Build()
			if err != nil {
				t.Fatalf("build: unexpected error <%v>", err)
			}

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/path", &bytes.Buffer{})

			// test request
			handler.ServeHTTP(response, request)
			if response.Body == nil {
				t.Fatalf("response has no body")
			}

		})
	}

}

func TestHandler_PermissionError(t *testing.T) {

	tt := []struct {
		name        string
		manifest    string
		permissions []string
		granted     bool
	}{
		{
			name:        "permission fulfilled",
			manifest:    `[ { "method": "GET", "path": "/path", "scope": [["A"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{"A"},
			granted:     true,
		},
		{
			name:        "missing permission",
			manifest:    `[ { "method": "GET", "path": "/path", "scope": [["A"]], "info": "info", "in": {}, "out": {} } ]`,
			permissions: []string{},
			granted:     false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			builder := &aicra.Builder{}
			if err := addDefaultTypes(builder); err != nil {
				t.Fatalf("unexpected error <%v>", err)
			}

			// add active permissions
			builder.WithContext(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := api.GetAuth(r.Context())
					if a == nil {
						t.Fatalf("cannot access api.Auth form request context")
					}

					a.Active = tc.permissions
					next.ServeHTTP(w, r)
				})
			})

			err := builder.Setup(strings.NewReader(tc.manifest))
			if err != nil {
				t.Fatalf("setup: unexpected error <%v>", err)
			}

			pathHandler := func(ctx context.Context) (*struct{}, api.Err) {
				return nil, api.ErrNotImplemented
			}

			if err := builder.Bind(http.MethodGet, "/path", pathHandler); err != nil {
				t.Fatalf("bind: unexpected error <%v>", err)
			}

			handler, err := builder.Build()
			if err != nil {
				t.Fatalf("build: unexpected error <%v>", err)
			}

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/path", &bytes.Buffer{})

			// test request
			handler.ServeHTTP(response, request)
			if response.Body == nil {
				t.Fatalf("response has no body")
			}
			type jsonResponse struct {
				Err api.Err `json:"error"`
			}
			var res jsonResponse
			err = json.Unmarshal(response.Body.Bytes(), &res)
			if err != nil {
				t.Fatalf("cannot unmarshal response: %s", err)
			}

			expectedError := api.ErrNotImplemented
			if !tc.granted {
				expectedError = api.ErrPermission
			}

			if res.Err.Code != expectedError.Code {
				t.Fatalf("expected error code %d got %d", expectedError.Code, res.Err.Code)
			}

		})
	}

}

func TestHandler_DynamicScope(t *testing.T) {
	tt := []struct {
		name        string
		manifest    string
		path        string
		handler     interface{}
		url         string
		body        string
		permissions []string
		granted     bool
	}{
		{
			name: "replace one granted",
			manifest: `[
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
			path:        "/path/{id}",
			handler:     func(context.Context, struct{ Input1 uint }) (*struct{}, api.Err) { return nil, api.ErrSuccess },
			url:         "/path/123",
			body:        ``,
			permissions: []string{"user[123]"},
			granted:     true,
		},
		{
			name: "replace one mismatch",
			manifest: `[
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
			path:        "/path/{id}",
			handler:     func(context.Context, struct{ Input1 uint }) (*struct{}, api.Err) { return nil, api.ErrSuccess },
			url:         "/path/666",
			body:        ``,
			permissions: []string{"user[123]"},
			granted:     false,
		},
		{
			name: "replace one valid dot separated",
			manifest: `[
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
			path:        "/path/{id}",
			handler:     func(context.Context, struct{ User uint }) (*struct{}, api.Err) { return nil, api.ErrSuccess },
			url:         "/path/123",
			body:        ``,
			permissions: []string{"prefix.user[123].suffix"},
			granted:     true,
		},
		{
			name: "replace two valid dot separated",
			manifest: `[
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
			path: "/prefix/{pid}/user/{uid}",
			handler: func(context.Context, struct {
				Prefix uint
				User   uint
			}) (*struct{}, api.Err) {
				return nil, api.ErrSuccess
			},
			url:         "/prefix/123/user/456",
			body:        ``,
			permissions: []string{"prefix[123].user[456].suffix"},
			granted:     true,
		},
		{
			name: "replace two invalid dot separated",
			manifest: `[
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
			path: "/prefix/{pid}/user/{uid}",
			handler: func(context.Context, struct {
				Prefix uint
				User   uint
			}) (*struct{}, api.Err) {
				return nil, api.ErrSuccess
			},
			url:         "/prefix/123/user/666",
			body:        ``,
			permissions: []string{"prefix[123].user[456].suffix"},
			granted:     false,
		},
		{
			name: "replace three valid dot separated",
			manifest: `[
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
			path: "/prefix/{pid}/user/{uid}/suffix/{sid}",
			handler: func(context.Context, struct {
				Prefix uint
				User   uint
				Suffix uint
			}) (*struct{}, api.Err) {
				return nil, api.ErrSuccess
			},
			url:         "/prefix/123/user/456/suffix/789",
			body:        ``,
			permissions: []string{"prefix[123].user[456].suffix[789]"},
			granted:     true,
		},
		{
			name: "replace three invalid dot separated",
			manifest: `[
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
			path: "/prefix/{pid}/user/{uid}/suffix/{sid}",
			handler: func(context.Context, struct {
				Prefix uint
				User   uint
				Suffix uint
			}) (*struct{}, api.Err) {
				return nil, api.ErrSuccess
			},
			url:         "/prefix/123/user/666/suffix/789",
			body:        ``,
			permissions: []string{"prefix[123].user[456].suffix[789]"},
			granted:     false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			builder := &aicra.Builder{}
			if err := addDefaultTypes(builder); err != nil {
				t.Fatalf("unexpected error <%v>", err)
			}

			// tester middleware (last executed)
			builder.WithContext(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := api.GetAuth(r.Context())
					if a == nil {
						t.Fatalf("cannot access api.Auth form request context")
					}
					if a.Granted() == tc.granted {
						return
					}
					if a.Granted() {
						t.Fatalf("unexpected granted auth")
					} else {
						t.Fatalf("expected granted auth")
					}
					next.ServeHTTP(w, r)
				})
			})

			// update permissions
			builder.WithContext(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := api.GetAuth(r.Context())
					if a == nil {
						t.Fatalf("cannot access api.Auth form request context")
					}
					a.Active = tc.permissions
					next.ServeHTTP(w, r)
				})
			})

			err := builder.Setup(strings.NewReader(tc.manifest))
			if err != nil {
				t.Fatalf("setup: unexpected error <%v>", err)
			}

			if err := builder.Bind(http.MethodPost, tc.path, tc.handler); err != nil {
				t.Fatalf("bind: unexpected error <%v>", err)
			}

			handler, err := builder.Build()
			if err != nil {
				t.Fatalf("build: unexpected error <%v>", err)
			}

			response := httptest.NewRecorder()
			body := strings.NewReader(tc.body)
			request := httptest.NewRequest(http.MethodPost, tc.url, body)

			// test request
			handler.ServeHTTP(response, request)
			if response.Body == nil {
				t.Fatalf("response has no body")
			}

		})
	}

}

func TestHandler_ServiceErrors(t *testing.T) {
	tt := []struct {
		name     string
		manifest string
		// handler
		hmethod, huri string
		hfn           interface{}
		// request
		method, url string
		contentType string
		body        string
		permissions []string
		err         api.Err
	}{
		// service match
		{
			name: "unknown service method",
			manifest: `[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {},
					"out": {}
				}
			]`,
			hmethod: http.MethodGet,
			huri:    "/",
			hfn: func(context.Context) api.Err {
				return api.ErrSuccess
			},
			method:      http.MethodPost,
			url:         "/",
			body:        ``,
			permissions: []string{},
			err:         api.ErrUnknownService,
		},
		{
			name: "unknown service path",
			manifest: `[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {},
					"out": {}
				}
			]`,
			hmethod: http.MethodGet,
			huri:    "/",
			hfn: func(context.Context) api.Err {
				return api.ErrSuccess
			},
			method:      http.MethodGet,
			url:         "/invalid",
			body:        ``,
			permissions: []string{},
			err:         api.ErrUnknownService,
		},
		{
			name: "valid empty service",
			manifest: `[
				{
					"method": "GET",
					"path": "/",
					"info": "info",
					"scope": [],
					"in":  {},
					"out": {}
				}
			]`,
			hmethod: http.MethodGet,
			huri:    "/",
			hfn: func(context.Context) api.Err {
				return api.ErrSuccess
			},
			method:      http.MethodGet,
			url:         "/",
			body:        ``,
			permissions: []string{},
			err:         api.ErrSuccess,
		},

		// invalid uri param -> unknown service
		{
			name: "invalid uri param",
			manifest: `[
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
			hmethod: http.MethodGet,
			huri:    "/a/{id}/b",
			hfn: func(context.Context, struct{ ID int }) api.Err {
				return api.ErrSuccess
			},
			method:      http.MethodGet,
			url:         "/a/invalid/b",
			body:        ``,
			permissions: []string{},
			err:         api.ErrUnknownService,
		},

		// query param
		{
			name: "missing query param",
			manifest: `[
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
			hmethod: http.MethodGet,
			huri:    "/",
			hfn: func(context.Context, struct{ ID int }) api.Err {
				return api.ErrSuccess
			},
			method:      http.MethodGet,
			url:         "/",
			body:        ``,
			permissions: []string{},
			err:         api.ErrMissingParam,
		},
		{
			name: "invalid query param",
			manifest: `[
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
			hmethod: http.MethodGet,
			huri:    "/a",
			hfn: func(context.Context, struct{ ID int }) api.Err {
				return api.ErrSuccess
			},
			method:      http.MethodGet,
			url:         "/a?id=abc",
			body:        ``,
			permissions: []string{},
			err:         api.ErrInvalidParam,
		},
		{
			name: "invalid query multi param",
			manifest: `[
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
			hmethod: http.MethodGet,
			huri:    "/a",
			hfn: func(context.Context, struct{ ID int }) api.Err {
				return api.ErrSuccess
			},
			method:      http.MethodGet,
			url:         "/a?id=123&id=456",
			body:        ``,
			permissions: []string{},
			err:         api.ErrInvalidParam,
		},
		{
			name: "valid query param",
			manifest: `[
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
			hmethod: http.MethodGet,
			huri:    "/a",
			hfn: func(context.Context, struct{ ID int }) api.Err {
				return api.ErrSuccess
			},
			method:      http.MethodGet,
			url:         "/a?id=123",
			body:        ``,
			permissions: []string{},
			err:         api.ErrSuccess,
		},

		// json param
		{
			name: "missing json param",
			manifest: `[
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
			hmethod: http.MethodPost,
			huri:    "/",
			hfn: func(context.Context, struct{ ID int }) api.Err {
				return api.ErrSuccess
			},
			contentType: "application/json",
			method:      http.MethodPost,
			url:         "/",
			body:        ``,
			permissions: []string{},
			err:         api.ErrMissingParam,
		},
		{
			name: "invalid json param",
			manifest: `[
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
			hmethod: http.MethodPost,
			huri:    "/",
			hfn: func(context.Context, struct{ ID int }) api.Err {
				return api.ErrSuccess
			},
			contentType: "application/json",
			method:      http.MethodPost,
			url:         "/",
			body:        `{ "id": "invalid type" }`,
			permissions: []string{},
			err:         api.ErrInvalidParam,
		},
		{
			name: "valid json param",
			manifest: `[
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
			hmethod: http.MethodPost,
			huri:    "/",
			hfn: func(context.Context, struct{ ID int }) api.Err {
				return api.ErrSuccess
			},
			contentType: "application/json",
			method:      http.MethodPost,
			url:         "/",
			body:        `{ "id": 123 }`,
			permissions: []string{},
			err:         api.ErrSuccess,
		},

		// urlencoded param
		{
			name: "missing urlencoded param",
			manifest: `[
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
			hmethod: http.MethodPost,
			huri:    "/",
			hfn: func(context.Context, struct{ ID int }) api.Err {
				return api.ErrSuccess
			},
			contentType: "application/x-www-form-urlencoded",
			method:      http.MethodPost,
			url:         "/",
			body:        ``,
			permissions: []string{},
			err:         api.ErrMissingParam,
		},
		{
			name: "invalid urlencoded param",
			manifest: `[
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
			hmethod: http.MethodPost,
			huri:    "/",
			hfn: func(context.Context, struct{ ID int }) api.Err {
				return api.ErrSuccess
			},
			contentType: "application/x-www-form-urlencoded",
			method:      http.MethodPost,
			url:         "/",
			body:        `id=abc`,
			permissions: []string{},
			err:         api.ErrInvalidParam,
		},
		{
			name: "valid urlencoded param",
			manifest: `[
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
			hmethod: http.MethodPost,
			huri:    "/",
			hfn: func(context.Context, struct{ ID int }) api.Err {
				return api.ErrSuccess
			},
			contentType: "application/x-www-form-urlencoded",
			method:      http.MethodPost,
			url:         "/",
			body:        `id=123`,
			permissions: []string{},
			err:         api.ErrSuccess,
		},

		// formdata param
		{
			name: "missing multipart param",
			manifest: `[
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
			hmethod: http.MethodPost,
			huri:    "/",
			hfn: func(context.Context, struct{ ID int }) api.Err {
				return api.ErrSuccess
			},
			contentType: "multipart/form-data; boundary=xxx",
			method:      http.MethodPost,
			url:         "/",
			body:        ``,
			permissions: []string{},
			err:         api.ErrMissingParam,
		},
		{
			name: "invalid multipart param",
			manifest: `[
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
			hmethod: http.MethodPost,
			huri:    "/",
			hfn: func(context.Context, struct{ ID int }) api.Err {
				return api.ErrSuccess
			},
			contentType: "multipart/form-data; boundary=xxx",
			method:      http.MethodPost,
			url:         "/",
			body: `--xxx
Content-Disposition: form-data; name="id"

abc
--xxx--`,
			permissions: []string{},
			err:         api.ErrInvalidParam,
		},
		{
			name: "valid multipart param",
			manifest: `[
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
			hmethod: http.MethodPost,
			huri:    "/",
			hfn: func(context.Context, struct{ ID int }) api.Err {
				return api.ErrSuccess
			},
			contentType: "multipart/form-data; boundary=xxx",
			method:      http.MethodPost,
			url:         "/",
			body: `--xxx
Content-Disposition: form-data; name="id"

123
--xxx--`,
			permissions: []string{},
			err:         api.ErrSuccess,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			builder := &aicra.Builder{}
			if err := addDefaultTypes(builder); err != nil {
				t.Fatalf("unexpected error <%v>", err)
			}

			err := builder.Setup(strings.NewReader(tc.manifest))
			if err != nil {
				t.Fatalf("setup: unexpected error <%v>", err)
			}

			if err := builder.Bind(tc.hmethod, tc.huri, tc.hfn); err != nil {
				t.Fatalf("bind: unexpected error <%v>", err)
			}

			handler, err := builder.Build()
			if err != nil {
				t.Fatalf("build: unexpected error <%v>", err)
			}

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
			if response.Body == nil {
				t.Fatalf("response has no body")
			}

			jsonErr, err := json.Marshal(tc.err)
			if err != nil {
				t.Fatalf("cannot marshal expected error: %v", err)
			}
			jsonExpected := fmt.Sprintf(`{"error":%s}`, jsonErr)
			if response.Body.String() != jsonExpected {
				t.Fatalf("invalid response:\n- actual: %s\n- expect: %s\n", printEscaped(response.Body.String()), printEscaped(jsonExpected))
			}
		})
	}
}
