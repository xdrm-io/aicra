package aicra

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/aicra/validator"
)

const baseConf = `` +
	/**/ `{` +
	/*	*/ `"package": "pkg",` +
	/*	*/ `"imports": {},` +
	/*	*/ `"validators": {` +
	/*		*/ `"string": { "use": "builtin.String", "as": "string" },` +
	/*		*/ `"uint": { "use": "builtin.Uint", "as": "uint" }` +
	/*	*/ `},` +
	/*	*/ `"endpoints": []` +
	/**/ `}`

func TestBuilder_Build(t *testing.T) {
	validators := config.Validators{
		"string": validator.Wrap[string](new(validator.String)),
		"uint":   validator.Wrap[uint](new(validator.Uint)),
	}

	tt := []struct {
		name       string
		setup      func(*testing.T, *Builder)
		validators config.Validators

		err  error
		want func(*testing.T, Handler)
	}{
		{
			name:  "nil config",
			setup: func(t *testing.T, b *Builder) {},
			err:   ErrNotSetup,
		},
		{
			name: "nil validators",
			setup: func(t *testing.T, b *Builder) {
				require.NoError(t, b.Setup(strings.NewReader(baseConf)))
			},
			err: ErrNilValidators,
		},
		{
			name: "default uri limit",
			setup: func(t *testing.T, b *Builder) {
				require.NoError(t, b.Setup(strings.NewReader(baseConf)))
			},
			validators: validators,
			err:        nil,
			want: func(t *testing.T, h Handler) {
				require.EqualValues(t, DefaultURILimit, h.uriLimit)
			},
		},
		{
			name: "default body limit",
			setup: func(t *testing.T, b *Builder) {
				require.NoError(t, b.Setup(strings.NewReader(baseConf)))
			},
			validators: validators,
			err:        nil,
			want: func(t *testing.T, h Handler) {
				require.EqualValues(t, DefaultBodyLimit, h.bodyLimit)
			},
		},
		{
			name: "custom uri limit",
			setup: func(t *testing.T, b *Builder) {
				require.NoError(t, b.Setup(strings.NewReader(baseConf)))
				b.SetURILimit(42)
			},
			validators: validators,
			err:        nil,
			want: func(t *testing.T, h Handler) {
				require.EqualValues(t, 42, h.uriLimit)
			},
		},
		{
			name: "custom body limit",
			setup: func(t *testing.T, b *Builder) {
				require.NoError(t, b.Setup(strings.NewReader(baseConf)))
				b.SetBodyLimit(42)
			},
			validators: validators,
			err:        nil,
			want: func(t *testing.T, h Handler) {
				require.EqualValues(t, 42, h.bodyLimit)
			},
		},

		{
			name: "missing handler",
			setup: func(t *testing.T, b *Builder) {
				require.NoError(t, b.Setup(strings.NewReader(baseConf)))
				b.conf.Endpoints = []*config.Endpoint{
					{Method: "GET", Pattern: "/path"},
				}
			},
			validators: validators,
			err:        ErrMissingHandler,
		},

		{
			name:       "path collision",
			validators: validators,
			setup: func(t *testing.T, b *Builder) {
				require.NoError(t, b.Setup(strings.NewReader(baseConf)))
				b.conf.Endpoints = []*config.Endpoint{
					{Method: "GET", Pattern: "/path/abc"},
					{
						Method: "GET", Pattern: "/path/{var}",
						Input: map[string]*config.Parameter{
							"{var}": {ValidatorName: "string", ValidatorParams: []string{"3"}},
						},
						Captures: []*config.BraceCapture{
							{Index: 1, Name: "var"},
						},
					},
				}
				b.handlers = []*serviceHandler{
					{Method: "GET", Path: "/path/abc", fn: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})},
					{Method: "GET", Path: "/path/{var}", fn: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})},
				}
			},
			err: config.ErrPatternCollision,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			builder := &Builder{}
			require.NotNil(t, tc.setup, "setup function must be defined")
			tc.setup(t, builder)

			handler, err := builder.Build(tc.validators)
			if tc.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			cast, ok := handler.(Handler)
			require.Truef(t, ok, "handler must be a aicra.Handler")
			tc.want(t, cast)
		})
	}
}

func TestBuilder_SetupTwice(t *testing.T) {
	t.Parallel()

	builder := &Builder{}
	err := builder.Setup(strings.NewReader(baseConf))
	require.NoError(t, err)
	// double Setup() must fail
	err = builder.Setup(strings.NewReader(baseConf))
	require.ErrorIs(t, err, ErrAlreadySetup)
}

func TestBuilder_BindBeforeSetup(t *testing.T) {
	t.Parallel()

	builder := &Builder{}

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// binding before Setup() must fail
	err := builder.Bind(http.MethodGet, "/path", fn)
	require.ErrorIs(t, err, ErrNotSetup)
}

func TestBuilder_Bind(t *testing.T) {
	tt := []struct {
		name         string
		endpoints    []config.Endpoint
		method, path string
		twice        bool
		fn           http.HandlerFunc
		err          error
	}{
		{
			name:   "nil handler",
			method: "GET", path: "/path",
			fn:  nil,
			err: ErrNilHandler,
		},
		{
			name:   "bind unknown endpoint",
			method: "GET", path: "/path",
			fn:  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			err: ErrUnknownService,
		},
		{
			name: "method mismatch",
			endpoints: []config.Endpoint{
				{Method: "GET", Pattern: "/path"},
			},
			method: "POST", path: "/path",
			fn:  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			err: ErrUnknownService,
		},
		{
			name: "path mismatch",
			endpoints: []config.Endpoint{
				{Method: "GET", Pattern: "/a/b/c"},
			},
			method: "GET", path: "/a/b/x",
			fn:  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			err: ErrUnknownService,
		},
		{
			name: "path mismatch with variable",
			endpoints: []config.Endpoint{
				{Method: "GET", Pattern: "/a/{var1}/c"},
			},
			method: "GET", path: "/a/{var2}/c",
			fn:  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			err: ErrUnknownService,
		},
		{
			name: "match",
			endpoints: []config.Endpoint{
				{Method: "GET", Pattern: "/a/b/c"},
			},
			method: "GET", path: "/a/b/c",
			fn:  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			err: nil,
		},
		{
			name: "match with variable",
			endpoints: []config.Endpoint{
				{Method: "GET", Pattern: "/a/{var}/c"},
			},
			method: "GET", path: "/a/{var}/c",
			fn:  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			err: nil,
		},

		{
			name: "bind twice",
			endpoints: []config.Endpoint{
				{Method: "GET", Pattern: "/a/b/c"},
			},
			method: "GET", path: "/a/b/c",
			twice: true,
			fn:    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			err:   ErrAlreadyBound,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			builder := &Builder{}
			err := builder.Setup(strings.NewReader(baseConf))
			require.NoError(t, err)
			for _, endpoint := range tc.endpoints {
				builder.conf.Endpoints = append(builder.conf.Endpoints, &endpoint)
			}
			if tc.twice {
				err = builder.Bind(tc.method, tc.path, tc.fn)
				require.NoError(t, err)
			}

			err = builder.Bind(tc.method, tc.path, tc.fn)
			if tc.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestBuilder_With(t *testing.T) {
	mwOk := func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	}
	mwCreated := func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		})
	}

	tt := []struct {
		name        string
		middlewares []func(http.Handler) http.Handler

		want []func(http.Handler) http.Handler
	}{
		{
			name:        "nil",
			middlewares: []func(http.Handler) http.Handler{nil},
		},
		{
			name:        "first allocates",
			middlewares: []func(http.Handler) http.Handler{mwOk},
			want:        []func(http.Handler) http.Handler{mwOk},
		},
		{
			name:        "2 in order",
			middlewares: []func(http.Handler) http.Handler{mwOk, mwCreated},
			want:        []func(http.Handler) http.Handler{mwOk, mwCreated},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			builder := &Builder{}
			for _, mw := range tc.middlewares {
				builder.With(mw)
			}

			require.Lenf(t, builder.middlewares, len(tc.want), "middlewares count mismatch")
			for i, mw := range tc.want {
				actual, expect := httptest.NewRecorder(), httptest.NewRecorder()
				builder.middlewares[i](nil).ServeHTTP(actual, nil)
				mw(nil).ServeHTTP(expect, nil)

				require.Equalf(t, expect.Code, actual.Code, "middleware #%d mismatch", i)
			}
		})
	}
}
func TestBuilder_WithContext(t *testing.T) {
	mwOk := func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	}
	mwCreated := func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		})
	}

	tt := []struct {
		name        string
		middlewares []func(http.Handler) http.Handler

		want []func(http.Handler) http.Handler
	}{
		{
			name:        "nil",
			middlewares: []func(http.Handler) http.Handler{nil},
		},
		{
			name:        "first allocates",
			middlewares: []func(http.Handler) http.Handler{mwOk},
			want:        []func(http.Handler) http.Handler{mwOk},
		},
		{
			name:        "2 in order",
			middlewares: []func(http.Handler) http.Handler{mwOk, mwCreated},
			want:        []func(http.Handler) http.Handler{mwOk, mwCreated},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			builder := &Builder{}
			for _, mw := range tc.middlewares {
				builder.WithContext(mw)
			}

			require.Lenf(t, builder.ctxMiddlewares, len(tc.want), "middlewares count mismatch")
			for i, mw := range tc.want {
				actual, expect := httptest.NewRecorder(), httptest.NewRecorder()
				builder.ctxMiddlewares[i](nil).ServeHTTP(actual, nil)
				mw(nil).ServeHTTP(expect, nil)

				require.Equalf(t, expect.Code, actual.Code, "middleware #%d mismatch", i)
			}
		})
	}
}
