package aicra

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/xdrm-io/aicra/internal/dynfunc"
	"github.com/xdrm-io/aicra/validator"
)

func addBuiltinTypes(b *Builder) error {
	inputTypes := []validator.Type{
		validator.AnyType{},
		validator.BoolType{},
		validator.FloatType{},
		validator.IntType{},
		validator.StringType{},
		validator.UintType{},
	}
	outputTypes := map[string]interface{}{
		"any":    interface{}(nil),
		"bool":   true,
		"float":  float64(2),
		"int":    int(0),
		"string": "",
		"uint":   uint(0),
	}

	for _, t := range inputTypes {
		if err := b.Input(t); err != nil {
			return err
		}
	}
	for k, v := range outputTypes {
		if err := b.Output(k, v); err != nil {
			return err
		}
	}
	return nil
}

func TestAddInputType(t *testing.T) {
	t.Parallel()

	builder := &Builder{}
	err := builder.Input(validator.BoolType{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = builder.Setup(strings.NewReader("[]"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = builder.Input(validator.FloatType{})
	if err != errLateType {
		t.Fatalf("expected <%v> got <%v>", errLateType, err)
	}
}
func TestAddOutputType(t *testing.T) {
	t.Parallel()

	builder := &Builder{}
	err := builder.Output("bool", true)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = builder.Setup(strings.NewReader("[]"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = builder.Output("bool", true)
	if err != errLateType {
		t.Fatalf("expected <%v> got <%v>", errLateType, err)
	}
}

func TestNilResponder(t *testing.T) {
	t.Parallel()

	builder := &Builder{}
	err := builder.RespondWith(nil)
	if !errors.Is(err, errNilResponder) {
		t.Fatalf("expected %q, got %v", errNilResponder.Error(), err)
	}
}
func TestNonNilResponder(t *testing.T) {
	t.Parallel()

	builder := &Builder{}
	err := builder.RespondWith(DefaultResponder)
	if !errors.Is(err, nil) {
		t.Fatalf("unexpected error %s", err)
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

	fn := func(context.Context, struct{}) (*struct{}, error) {
		return nil, nil
	}

	// binding before Setup() must fail
	err := Bind(builder, http.MethodGet, "/path", fn)
	if err != errNotSetup {
		t.Fatalf("expected error %v, got %v", errNotSetup, err)
	}
}

func TestBindUnknownService(t *testing.T) {
	t.Parallel()

	fn := func(context.Context, struct{}) (*struct{}, error) {
		return nil, nil
	}

	builder := &Builder{}
	err := builder.Setup(strings.NewReader("[]"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = Bind(builder, http.MethodGet, "/path", fn)
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
	fn := func(context.Context, struct{ Unexpected int }) (*struct{}, error) {
		return nil, nil
	}
	err = Bind(builder, http.MethodGet, "/path", fn)
	if !errors.Is(err, dynfunc.ErrUnexpectedFields) {
		t.Fatalf("expected a dynfunc.Err got %v", err)
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
	fn := func(context.Context, struct{}) (*struct{}, error) {
		return nil, nil
	}
	err = Bind(builder, http.MethodGet, "/path", fn)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	_, err = builder.Build()
	if !errors.Is(err, errMissingHandler) {
		t.Fatalf("expected a %v error, got %v", errMissingHandler, err)
	}
}

func bind[Req, Res any](method, path string, fn HandlerFunc[Req, Res]) func(*Builder) error {
	return func(b *Builder) error {
		return Bind(b, method, path, fn)
	}
}

func TestBind(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		conf     string
		binder   func(*Builder) error
		bindErr  error
		buildErr error
	}{
		{
			name:     "none required none provided",
			conf:     "[]",
			binder:   nil,
			buildErr: nil,
		},
		{
			name:     "none required 1 provided",
			conf:     "[]",
			binder:   bind("", "", func(context.Context, struct{}) (*struct{}, error) { return nil, nil }),
			bindErr:  errUnknownService,
			buildErr: nil,
		},
		{
			name: "1 required none provided",
			conf: `[
				{
					"method": "GET",
					"path": "/path",
					"scope": [[]],
					"info": "info",
					"in": {},
					"out": {}
				}
			]`,
			binder:   nil,
			buildErr: errMissingHandler,
		},
		{
			name: "1 required wrong method provided",
			conf: `[
				{
					"method": "GET",
					"path": "/path",
					"scope": [[]],
					"info": "info",
					"in": {},
					"out": {}
				}
			]`,
			binder:   bind("POST", "/path", func(context.Context, struct{}) (*struct{}, error) { return nil, nil }),
			bindErr:  errUnknownService,
			buildErr: errMissingHandler,
		},
		{
			name: "1 required wrong path provided",
			conf: `[
				{
					"method": "GET",
					"path": "/path",
					"scope": [[]],
					"info": "info",
					"in": {},
					"out": {}
				}
			]`,
			binder:   bind("POST", "/paths", func(context.Context, struct{}) (*struct{}, error) { return nil, nil }),
			bindErr:  errUnknownService,
			buildErr: errMissingHandler,
		},
		{
			name: "1 required valid provided",
			conf: `[
				{
					"method": "GET",
					"path": "/path",
					"scope": [[]],
					"info": "info",
					"in": {},
					"out": {}
				}
			]`,
			binder:   bind("GET", "/path", func(context.Context, struct{}) (*struct{}, error) { return nil, nil }),
			bindErr:  nil,
			buildErr: nil,
		},
		{
			name: "1 required with int",
			conf: `[
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
			binder:   bind("GET", "/path", func(context.Context, struct{ Name int }) (*struct{}, error) { return nil, nil }),
			bindErr:  nil,
			buildErr: nil,
		},
		{
			name: "1 required with uint",
			conf: `[
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
			binder:   bind("GET", "/path", func(context.Context, struct{ Name uint }) (*struct{}, error) { return nil, nil }),
			bindErr:  nil,
			buildErr: nil,
		},
		{
			name: "1 required with string",
			conf: `[
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
			binder:   bind("GET", "/path", func(context.Context, struct{ Name string }) (*struct{}, error) { return nil, nil }),
			bindErr:  nil,
			buildErr: nil,
		},
		{
			name: "1 required with bool",
			conf: `[
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
			binder:   bind("GET", "/path", func(context.Context, struct{ Name bool }) (*struct{}, error) { return nil, nil }),
			bindErr:  nil,
			buildErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			builder := &Builder{}
			if err := addBuiltinTypes(builder); err != nil {
				t.Fatalf("add built-in types: %s", err)
			}

			err := builder.Setup(strings.NewReader(tc.conf))
			if err != nil {
				t.Fatalf("setup: unexpected error <%v>", err)
			}

			if tc.binder != nil {
				err := tc.binder(builder)
				if !errors.Is(err, tc.bindErr) {
					t.Fatalf("invalid bind error\nactual: %v\nexpect: %v", err, tc.bindErr)
				}
			}

			_, err = builder.Build()
			if !errors.Is(err, tc.buildErr) {
				t.Fatalf("invalid build error\nactual: %v\nexpect: %v", err, tc.buildErr)
			}

		})
	}
}
