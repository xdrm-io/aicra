package aicra

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/datatype/builtin"
)

func addBuiltinTypes(b *Builder) error {
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

func TestAddType(t *testing.T) {
	builder := &Builder{}
	err := builder.AddType(builtin.BoolDataType{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = builder.Setup(strings.NewReader("[]"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = builder.AddType(builtin.FloatDataType{})
	if err != errLateType {
		t.Fatalf("expected <%v> got <%v>", errLateType, err)
	}
}

func TestUse(t *testing.T) {
	builder := &Builder{}
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
		builder.Use(middleware)
	}

	config := strings.NewReader(`[ { "method": "GET", "path": "/path", "scope": [[]], "info": "info", "in": {}, "out": {} } ]`)
	err := builder.Setup(config)
	if err != nil {
		t.Fatalf("setup: unexpected error <%v>", err)
	}

	pathHandler := func(ctx api.Ctx) (*struct{}, api.Err) {
		// write value from middlewares into response
		value := ctx.Req.Context().Value(key)
		if value == nil {
			t.Fatalf("nothing found in context")
		}
		cast, ok := value.(int)
		if !ok {
			t.Fatalf("cannot cast context data to int")
		}
		// write to response
		ctx.Res.Write([]byte(fmt.Sprintf("#%d#", cast)))

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

func TestBind(t *testing.T) {
	tcases := []struct {
		Name          string
		Config        string
		HandlerMethod string
		HandlerPath   string
		HandlerFn     interface{} // not bound if nil
		BindErr       error
		BuildErr      error
	}{
		{
			Name:          "none required none provided",
			Config:        "[]",
			HandlerMethod: "",
			HandlerPath:   "",
			HandlerFn:     nil,
			BindErr:       nil,
			BuildErr:      nil,
		},
		{
			Name:          "none required 1 provided",
			Config:        "[]",
			HandlerMethod: "",
			HandlerPath:   "",
			HandlerFn:     func() (*struct{}, api.Err) { return nil, api.ErrSuccess },
			BindErr:       errUnknownService,
			BuildErr:      nil,
		},
		{
			Name: "1 required none provided",
			Config: `[
				{
					"method": "GET",
					"path": "/path",
					"scope": [[]],
					"info": "info",
					"in": {},
					"out": {}
				}
			]`,
			HandlerMethod: "",
			HandlerPath:   "",
			HandlerFn:     nil,
			BindErr:       nil,
			BuildErr:      errMissingHandler,
		},
		{
			Name: "1 required wrong method provided",
			Config: `[
				{
					"method": "GET",
					"path": "/path",
					"scope": [[]],
					"info": "info",
					"in": {},
					"out": {}
				}
			]`,
			HandlerMethod: http.MethodPost,
			HandlerPath:   "/path",
			HandlerFn:     func() (*struct{}, api.Err) { return nil, api.ErrSuccess },
			BindErr:       errUnknownService,
			BuildErr:      errMissingHandler,
		},
		{
			Name: "1 required wrong path provided",
			Config: `[
				{
					"method": "GET",
					"path": "/path",
					"scope": [[]],
					"info": "info",
					"in": {},
					"out": {}
				}
			]`,
			HandlerMethod: http.MethodGet,
			HandlerPath:   "/paths",
			HandlerFn:     func() (*struct{}, api.Err) { return nil, api.ErrSuccess },
			BindErr:       errUnknownService,
			BuildErr:      errMissingHandler,
		},
		{
			Name: "1 required valid provided",
			Config: `[
				{
					"method": "GET",
					"path": "/path",
					"scope": [[]],
					"info": "info",
					"in": {},
					"out": {}
				}
			]`,
			HandlerMethod: http.MethodGet,
			HandlerPath:   "/path",
			HandlerFn:     func() (*struct{}, api.Err) { return nil, api.ErrSuccess },
			BindErr:       nil,
			BuildErr:      nil,
		},
		{
			Name: "1 required with int",
			Config: `[
				{
					"method": "GET",
					"path": "/path",
					"scope": [[]],
					"info": "info",
					"in": {
						"id": { "info": "info", "type": "int", "name": "Name" }
					},
					"out": {}
				}
			]`,
			HandlerMethod: http.MethodGet,
			HandlerPath:   "/path",
			HandlerFn:     func(struct{ Name int }) (*struct{}, api.Err) { return nil, api.ErrSuccess },
			BindErr:       nil,
			BuildErr:      nil,
		},
		{
			Name: "1 required with uint",
			Config: `[
				{
					"method": "GET",
					"path": "/path",
					"scope": [[]],
					"info": "info",
					"in": {
						"id": { "info": "info", "type": "uint", "name": "Name" }
					},
					"out": {}
				}
			]`,
			HandlerMethod: http.MethodGet,
			HandlerPath:   "/path",
			HandlerFn:     func(struct{ Name uint }) (*struct{}, api.Err) { return nil, api.ErrSuccess },
			BindErr:       nil,
			BuildErr:      nil,
		},
		{
			Name: "1 required with string",
			Config: `[
				{
					"method": "GET",
					"path": "/path",
					"scope": [[]],
					"info": "info",
					"in": {
						"id": { "info": "info", "type": "string", "name": "Name" }
					},
					"out": {}
				}
			]`,
			HandlerMethod: http.MethodGet,
			HandlerPath:   "/path",
			HandlerFn:     func(struct{ Name string }) (*struct{}, api.Err) { return nil, api.ErrSuccess },
			BindErr:       nil,
			BuildErr:      nil,
		},
		{
			Name: "1 required with bool",
			Config: `[
				{
					"method": "GET",
					"path": "/path",
					"scope": [[]],
					"info": "info",
					"in": {
						"id": { "info": "info", "type": "bool", "name": "Name" }
					},
					"out": {}
				}
			]`,
			HandlerMethod: http.MethodGet,
			HandlerPath:   "/path",
			HandlerFn:     func(struct{ Name bool }) (*struct{}, api.Err) { return nil, api.ErrSuccess },
			BindErr:       nil,
			BuildErr:      nil,
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.Name, func(t *testing.T) {
			t.Parallel()

			builder := &Builder{}

			if err := addBuiltinTypes(builder); err != nil {
				t.Fatalf("add built-in types: %s", err)
			}

			err := builder.Setup(strings.NewReader(tcase.Config))
			if err != nil {
				t.Fatalf("setup: unexpected error <%v>", err)
			}

			if tcase.HandlerFn != nil {
				err := builder.Bind(tcase.HandlerMethod, tcase.HandlerPath, tcase.HandlerFn)
				if !errors.Is(err, tcase.BindErr) {
					t.Fatalf("bind: expected <%v> got <%v>", tcase.BindErr, err)
				}
			}

			_, err = builder.Build()
			if !errors.Is(err, tcase.BuildErr) {
				t.Fatalf("build: expected <%v> got <%v>", tcase.BuildErr, err)
			}

		})
	}
}
