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

func addBuiltinTypes(b *aicra.Builder) error {
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

func TestWith(t *testing.T) {
	builder := &aicra.Builder{}
	if err := addBuiltinTypes(builder); err != nil {
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

func TestWithAuth(t *testing.T) {

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
			if err := addBuiltinTypes(builder); err != nil {
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

func TestPermissionError(t *testing.T) {

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
			if err := addBuiltinTypes(builder); err != nil {
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

func TestDynamicScope(t *testing.T) {
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
			if err := addBuiltinTypes(builder); err != nil {
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
