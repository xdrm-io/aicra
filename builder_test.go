package aicra

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

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

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// binding before Setup() must fail
	err := builder.Bind(http.MethodGet, "/path", fn)
	if err != errNotSetup {
		t.Fatalf("expected error %v, got %v", errNotSetup, err)
	}
}

func TestBindUnknownService(t *testing.T) {
	t.Parallel()

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	builder := &Builder{}
	err := builder.Setup(strings.NewReader("[]"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = builder.Bind(http.MethodGet, "/path", fn)
	if !errors.Is(err, errUnknownService) {
		t.Fatalf("expected error %v, got %v", errUnknownService, err)
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
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	err = builder.Bind(http.MethodGet, "/path", fn)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	_, err = builder.Build()
	if !errors.Is(err, errMissingHandler) {
		t.Fatalf("expected a %v error, got %v", errMissingHandler, err)
	}
}

func bind(method, path string, fn http.HandlerFunc) func(*Builder) error {
	return func(b *Builder) error {
		return b.Bind(method, path, fn)
	}
}

func TestBind(t *testing.T) {
	t.Parallel()

	noOpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

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
			binder:   bind("", "", noOpHandler),
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
			binder:   bind("POST", "/path", noOpHandler),
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
			binder:   bind("POST", "/paths", noOpHandler),
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
			binder:   bind("GET", "/path", noOpHandler),
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
			binder:   bind("GET", "/path", noOpHandler),
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
			binder:   bind("GET", "/path", noOpHandler),
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
			binder:   bind("GET", "/path", noOpHandler),
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
			binder:   bind("GET", "/path", noOpHandler),
			bindErr:  nil,
			buildErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			builder := &Builder{}
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
