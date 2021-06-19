package aicra_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.xdrm.io/go/aicra"
	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/datatype/builtin"
)

func addBuiltinTypes(b *aicra.Builder) error {
	if err := b.AddType(builtin.AnyDataType{}); err != nil {
		return err
	}
	if err := b.AddType(builtin.BoolDataType{}); err != nil {
		return err
	}
	if err := b.AddType(builtin.FloatDataType{}); err != nil {
		return err
	}
	if err := b.AddType(builtin.IntDataType{}); err != nil {
		return err
	}
	if err := b.AddType(builtin.StringDataType{}); err != nil {
		return err
	}
	if err := b.AddType(builtin.UintDataType{}); err != nil {
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

	middleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			newr := r

			// first time -> store 1
			value := r.Context().Value(key)
			if value == nil {
				newr = r.WithContext(context.WithValue(r.Context(), key, int(1)))
				next(w, newr)
				return
			}

			// get value and increment
			cast, ok := value.(int)
			if !ok {
				t.Fatalf("value is not an int")
			}
			cast++
			newr = r.WithContext(context.WithValue(r.Context(), key, cast))
			next(w, newr)
		}
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

	pathHandler := func(ctx *api.Context) (*struct{}, api.Err) {
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
		ctx.ResponseWriter().Write([]byte(fmt.Sprintf("#%d#", cast)))

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
			builder.WithAuth(func(next api.AuthHandlerFunc) api.AuthHandlerFunc {
				return func(a api.Auth, w http.ResponseWriter, r *http.Request) {
					if a.Granted() == tc.granted {
						return
					}
					if a.Granted() {
						t.Fatalf("unexpected granted auth")
					} else {
						t.Fatalf("expected granted auth")
					}
				}
			})

			builder.WithAuth(func(next api.AuthHandlerFunc) api.AuthHandlerFunc {
				return func(a api.Auth, w http.ResponseWriter, r *http.Request) {
					a.Active = tc.permissions
					next(a, w, r)
				}
			})

			err := builder.Setup(strings.NewReader(tc.manifest))
			if err != nil {
				t.Fatalf("setup: unexpected error <%v>", err)
			}

			pathHandler := func(ctx *api.Context) (*struct{}, api.Err) {
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
			handler:     func(*api.Context, struct{ Input1 uint }) (*struct{}, api.Err) { return nil, api.ErrSuccess },
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
			handler:     func(*api.Context, struct{ Input1 uint }) (*struct{}, api.Err) { return nil, api.ErrSuccess },
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
			handler:     func(*api.Context, struct{ User uint }) (*struct{}, api.Err) { return nil, api.ErrSuccess },
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
			handler: func(*api.Context, struct {
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
			handler: func(*api.Context, struct {
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
			handler: func(*api.Context, struct {
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
			handler: func(*api.Context, struct {
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
			builder.WithAuth(func(next api.AuthHandlerFunc) api.AuthHandlerFunc {
				return func(a api.Auth, w http.ResponseWriter, r *http.Request) {
					if a.Granted() == tc.granted {
						return
					}
					if a.Granted() {
						t.Fatalf("unexpected granted auth")
					} else {
						t.Fatalf("expected granted auth")
					}
				}
			})

			// update permissions
			builder.WithAuth(func(next api.AuthHandlerFunc) api.AuthHandlerFunc {
				return func(a api.Auth, w http.ResponseWriter, r *http.Request) {
					a.Active = tc.permissions
					next(a, w, r)
				}
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
