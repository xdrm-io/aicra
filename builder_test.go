package aicra

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/validator/builtin"
)

func addBuiltinTypes(b *Builder) error {
	if err := b.Validate(builtin.AnyDataType{}); err != nil {
		return err
	}
	if err := b.Validate(builtin.BoolDataType{}); err != nil {
		return err
	}
	if err := b.Validate(builtin.FloatDataType{}); err != nil {
		return err
	}
	if err := b.Validate(builtin.IntDataType{}); err != nil {
		return err
	}
	if err := b.Validate(builtin.StringDataType{}); err != nil {
		return err
	}
	if err := b.Validate(builtin.UintDataType{}); err != nil {
		return err
	}
	return nil
}

func TestAddType(t *testing.T) {
	builder := &Builder{}
	err := builder.Validate(builtin.BoolDataType{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = builder.Setup(strings.NewReader("[]"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = builder.Validate(builtin.FloatDataType{})
	if err != errLateType {
		t.Fatalf("expected <%v> got <%v>", errLateType, err)
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
