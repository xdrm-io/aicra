package aicra

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/internal/config"
)

func TestSetupNoType(t *testing.T) {
	t.Parallel()

	builder := &Builder{}
	err := builder.Setup(strings.NewReader("[]"))
	require.NoError(t, err)
}

func TestSetupTwice(t *testing.T) {
	t.Parallel()

	builder := &Builder{}
	err := builder.Setup(strings.NewReader("[]"))
	require.NoError(t, err)
	// double Setup() must fail
	err = builder.Setup(strings.NewReader("[]"))
	require.ErrorIs(t, err, errAlreadySetup)
}

func TestBindBeforeSetup(t *testing.T) {
	t.Parallel()

	builder := &Builder{}

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// binding before Setup() must fail
	err := builder.Bind(http.MethodGet, "/path", fn)
	require.ErrorIs(t, err, errNotSetup)
}

func TestBindUnknownService(t *testing.T) {
	t.Parallel()

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	builder := &Builder{}
	err := builder.Setup(strings.NewReader("[]"))
	require.NoError(t, err)
	err = builder.Bind(http.MethodGet, "/path", fn)
	require.ErrorIs(t, err, errUnknownService)
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
	require.NoError(t, err)
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	err = builder.Bind(http.MethodGet, "/path", fn)
	require.NoError(t, err)

	_, err = builder.Build(nil)
	require.ErrorIs(t, err, errMissingHandler)
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
			tc := tc
			t.Parallel()

			builder := &Builder{}
			err := builder.Setup(strings.NewReader(tc.conf))
			require.NoError(t, err)

			if tc.binder != nil {
				err := tc.binder(builder)
				require.ErrorIs(t, err, tc.bindErr, "bind error")
			}

			_, err = builder.Build(config.Validators{})
			require.ErrorIs(t, err, tc.buildErr, "build error")
		})
	}
}
