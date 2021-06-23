package aicra

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/internal/dynfunc"
	"github.com/xdrm-io/aicra/validator"
)

func addBuiltinTypes(b *Builder) error {
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

func TestAddType(t *testing.T) {
	t.Parallel()

	builder := &Builder{}
	err := builder.Validate(validator.BoolType{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = builder.Setup(strings.NewReader("[]"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = builder.Validate(validator.FloatType{})
	if err != errLateType {
		t.Fatalf("expected <%v> got <%v>", errLateType, err)
	}
}

func TestSetupNoType(t *testing.T) {
	t.Parallel()

	builder := &Builder{}
	err := builder.Setup(strings.NewReader("[]"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}
func TestSetupTwice(t *testing.T) {
	t.Parallel()

	builder := &Builder{}
	err := builder.Setup(strings.NewReader("[]"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	// double Setup() must fail
	err = builder.Setup(strings.NewReader("[]"))
	if err != errAlreadySetup {
		t.Fatalf("expected error %v, got %v", errAlreadySetup, err)
	}
}

func TestBindBeforeSetup(t *testing.T) {
	t.Parallel()

	builder := &Builder{}
	// binding before Setup() must fail
	err := builder.Bind(http.MethodGet, "/path", func() {})
	if err != errNotSetup {
		t.Fatalf("expected error %v, got %v", errNotSetup, err)
	}
}

func TestBindUnknownService(t *testing.T) {
	t.Parallel()

	builder := &Builder{}
	err := builder.Setup(strings.NewReader("[]"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = builder.Bind(http.MethodGet, "/path", func() {})
	if !errors.Is(err, errUnknownService) {
		t.Fatalf("expected error %v, got %v", errUnknownService, err)
	}
}
func TestBindInvalidHandler(t *testing.T) {
	t.Parallel()

	builder := &Builder{}
	err := builder.Setup(strings.NewReader(`[
		{
			"method": "GET",
			"path": "/path",
			"scope": [[]],
			"info": "info",
			"in": {},
			"out": {}
		}
	]`))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = builder.Bind(http.MethodGet, "/path", func() {})

	if err == nil {
		t.Fatalf("expected an error")
	}

	if !errors.Is(err, dynfunc.ErrMissingHandlerContextArgument) {
		t.Fatalf("expected a dynfunc.Err got %v", err)
	}
}
func TestBindGet(t *testing.T) {
	t.Parallel()

	builder := &Builder{}
	err := builder.Setup(strings.NewReader(`[
		{
			"method": "GET",
			"path": "/path",
			"scope": [[]],
			"info": "info",
			"in": {},
			"out": {}
		},
		{
			"method": "POST",
			"path": "/path",
			"scope": [[]],
			"info": "info",
			"in": {},
			"out": {}
		},
		{
			"method": "PUT",
			"path": "/path",
			"scope": [[]],
			"info": "info",
			"in": {},
			"out": {}
		},
		{
			"method": "DELETE",
			"path": "/path",
			"scope": [[]],
			"info": "info",
			"in": {},
			"out": {}
		}
	]`))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	err = builder.Get("/path", func(context.Context) (*struct{}, api.Err) { return nil, api.ErrSuccess })
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	err = builder.Post("/path", func(context.Context) (*struct{}, api.Err) { return nil, api.ErrSuccess })
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	err = builder.Put("/path", func(context.Context) (*struct{}, api.Err) { return nil, api.ErrSuccess })
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	err = builder.Delete("/path", func(context.Context) (*struct{}, api.Err) { return nil, api.ErrSuccess })
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestUnhandledService(t *testing.T) {
	t.Parallel()

	builder := &Builder{}
	err := builder.Setup(strings.NewReader(`[
		{
			"method": "GET",
			"path": "/path",
			"scope": [[]],
			"info": "info",
			"in": {},
			"out": {}
		},
		{
			"method": "POST",
			"path": "/path",
			"scope": [[]],
			"info": "info",
			"in": {},
			"out": {}
		}
	]`))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	err = builder.Get("/path", func(context.Context) (*struct{}, api.Err) { return nil, api.ErrSuccess })
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	_, err = builder.Build()
	if !errors.Is(err, errMissingHandler) {
		t.Fatalf("expected a %v error, got %v", errMissingHandler, err)
	}
}
func TestBind(t *testing.T) {
	t.Parallel()

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
			HandlerFn:     func(context.Context) (*struct{}, api.Err) { return nil, api.ErrSuccess },
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
			HandlerFn:     func(context.Context) (*struct{}, api.Err) { return nil, api.ErrSuccess },
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
			HandlerFn:     func(context.Context) (*struct{}, api.Err) { return nil, api.ErrSuccess },
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
			HandlerFn:     func(context.Context) (*struct{}, api.Err) { return nil, api.ErrSuccess },
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
			HandlerFn:     func(context.Context, struct{ Name int }) (*struct{}, api.Err) { return nil, api.ErrSuccess },
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
			HandlerFn:     func(context.Context, struct{ Name uint }) (*struct{}, api.Err) { return nil, api.ErrSuccess },
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
			HandlerFn:     func(context.Context, struct{ Name string }) (*struct{}, api.Err) { return nil, api.ErrSuccess },
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
			HandlerFn:     func(context.Context, struct{ Name bool }) (*struct{}, api.Err) { return nil, api.ErrSuccess },
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
