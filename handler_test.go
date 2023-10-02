package aicra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/aicra/runtime"
)

func TestHandler_With(t *testing.T) {
	builder := &Builder{}

	err := builder.Setup(strings.NewReader(baseConf))
	require.NoError(t, err)

	builder.conf.Endpoints = []*config.Endpoint{
		{Method: "GET", Pattern: "/path", Fragments: []string{"path"}},
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
			require.True(t, ok, "cannot cast context data to int")
			cast++
			r = r.WithContext(context.WithValue(r.Context(), key, cast))
			next.ServeHTTP(w, r)
		})
	}

	// add middleware @n times
	for i := 0; i < n; i++ {
		builder.With(middleware)
	}

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// write value from middlewares into response
		value := r.Context().Value(key)
		require.NotNil(t, value, "cannot get context value")
		cast, ok := value.(int)
		require.True(t, ok, "cannot cast context data to int")
		// write to response
		w.Write([]byte(fmt.Sprintf("#%d#", cast)))
	})

	err = builder.Bind(http.MethodGet, "/path", fn)
	require.NoError(t, err)

	handler, err := builder.Build(config.Validators{})
	require.NoError(t, err)

	req, res := httptest.NewRequest("GET", "/path", nil), httptest.NewRecorder()

	// test request
	handler.ServeHTTP(res, req)
	require.NotNil(t, res.Body, "body")
	token := fmt.Sprintf("#%d#", n)
	require.Containsf(t, res.Body.String(), token, "expected %q to be in response %q", token, res.Body)
}

func TestHandler_Auth(t *testing.T) {
	tt := []struct {
		name        string
		endpoint    config.Endpoint
		permissions []string
		granted     bool
	}{
		{
			name:     "nil requirements",
			endpoint: config.Endpoint{Method: "GET", Pattern: "/path", Scope: nil},
			granted:  true,
		},
		{
			name:     "empty requirements",
			endpoint: config.Endpoint{Method: "GET", Pattern: "/path", Scope: [][]string{}},
			granted:  true,
		},
		{
			name:        "provide only requirement A",
			endpoint:    config.Endpoint{Method: "GET", Pattern: "/path", Scope: [][]string{{"A"}}},
			permissions: []string{"A"},
			granted:     true,
		},
		{
			name:        "missing requirement",
			endpoint:    config.Endpoint{Method: "GET", Pattern: "/path", Scope: [][]string{{"A"}}},
			permissions: []string{},
			granted:     false,
		},
		{
			name:        "missing requirements",
			endpoint:    config.Endpoint{Method: "GET", Pattern: "/path", Scope: [][]string{{"A", "B"}}},
			permissions: []string{},
			granted:     false,
		},
		{
			name:        "missing some requirements",
			endpoint:    config.Endpoint{Method: "GET", Pattern: "/path", Scope: [][]string{{"A", "B"}}},
			permissions: []string{"A"},
			granted:     false,
		},
		{
			name:        "provide requirements",
			endpoint:    config.Endpoint{Method: "GET", Pattern: "/path", Scope: [][]string{{"A", "B"}}},
			permissions: []string{"A", "B"},
			granted:     true,
		},
		{
			name:        "missing OR requirements",
			endpoint:    config.Endpoint{Method: "GET", Pattern: "/path", Scope: [][]string{{"A"}, {"B"}}},
			permissions: []string{"C"},
			granted:     false,
		},
		{
			name:        "provide 1 OR requirement",
			endpoint:    config.Endpoint{Method: "GET", Pattern: "/path", Scope: [][]string{{"A"}, {"B"}}},
			permissions: []string{"A"},
			granted:     true,
		},
		{
			name:        "provide both OR requirements",
			endpoint:    config.Endpoint{Method: "GET", Pattern: "/path", Scope: [][]string{{"A"}, {"B"}}},
			permissions: []string{"A", "B"},
			granted:     true,
		},
		{
			name:        "missing composite OR requirements",
			endpoint:    config.Endpoint{Method: "GET", Pattern: "/path", Scope: [][]string{{"A", "B"}, {"C", "D"}}},
			permissions: []string{},
			granted:     false,
		},
		{
			name:        "missing partial composite OR requirements",
			endpoint:    config.Endpoint{Method: "GET", Pattern: "/path", Scope: [][]string{{"A", "B"}, {"C", "D"}}},
			permissions: []string{"A", "C"},
			granted:     false,
		},
		{
			name:        "provide 1 composite OR requirement",
			endpoint:    config.Endpoint{Method: "GET", Pattern: "/path", Scope: [][]string{{"A", "B"}, {"C", "D"}}},
			permissions: []string{"A", "B", "C"},
			granted:     true,
		},
		{
			name:        "provide both composite OR requirements",
			endpoint:    config.Endpoint{Method: "GET", Pattern: "/path", Scope: [][]string{{"A", "B"}, {"C", "D"}}},
			permissions: []string{"A", "B", "C", "D"},
			granted:     true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			builder := &Builder{}

			// tester middleware (last executed)
			builder.WithContext(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := api.Extract(r.Context()).Auth
					require.NotNil(t, a, "cannot access api.Auth form request context")

					require.Equal(t, tc.granted, a.Granted(), "auth granted")
					next.ServeHTTP(w, r)
				})
			})

			builder.WithContext(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := api.Extract(r.Context()).Auth
					require.NotNil(t, a, "cannot access api.Auth form request context")

					a.Active = tc.permissions
					next.ServeHTTP(w, r)
				})
			})

			err := builder.Setup(strings.NewReader(baseConf))
			require.NoError(t, err)

			builder.conf.Endpoints = []*config.Endpoint{&tc.endpoint}

			pathHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				runtime.Respond(w, nil, api.ErrNotImplemented)
			})

			err = builder.Bind(http.MethodGet, "/path", pathHandler)
			require.NoError(t, err)

			handler, err := builder.Build(config.Validators{})
			require.NoError(t, err)

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/path", &bytes.Buffer{})

			// test request
			handler.ServeHTTP(response, request)
			require.NotNil(t, response.Body, "body")

		})
	}

}

func TestHandler_PermissionError(t *testing.T) {
	noOpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runtime.Respond(w, nil, api.ErrNotImplemented)
	})

	tt := []struct {
		name      string
		uri, body string
		endpoint  config.Endpoint

		method, path string
		permissions  []string
		err          api.Err
	}{
		{
			name: "permission fulfilled",
			uri:  "/path",
			endpoint: config.Endpoint{
				Method: "GET", Pattern: "/path", Scope: [][]string{{"A"}},
				Fragments: []string{"path"},
			},
			method: "GET", path: "/path",
			permissions: []string{"A"},
			err:         api.ErrNotImplemented,
		},
		{
			name: "missing permission",
			uri:  "/path",
			endpoint: config.Endpoint{
				Method: "GET", Pattern: "/path", Scope: [][]string{{"A"}},
				Fragments: []string{"path"},
			},
			method: "GET", path: "/path",
			permissions: []string{},
			err:         api.ErrForbidden,
		},
		// check that permission errors are raised:
		// AFTER uri params
		// BEFORE query and body params
		{
			name: "permission with wrong uri param",
			uri:  "/path/abc",
			endpoint: config.Endpoint{
				Method: "GET", Pattern: "/path/{uid}", Scope: [][]string{{"A"}},
				Input: map[string]*config.Parameter{
					"{uid}": {ValidatorName: "uint", Kind: config.KindURI},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "uid"},
				},
				Fragments: []string{"path", "{uid}"},
			},
			method: "GET", path: "/path/{uid}",
			permissions: []string{},
			err:         api.ErrUnknownService,
		},
		{
			name: "permission with wrong query param",
			uri:  "/path?uid=invalid-type",
			endpoint: config.Endpoint{
				Method: "GET", Pattern: "/path", Scope: [][]string{{"A"}},
				Input: map[string]*config.Parameter{
					"?uid": {ValidatorName: "uint", Kind: config.KindQuery},
				},
				Fragments: []string{"path"},
			},
			method: "GET", path: "/path",
			permissions: []string{},
			err:         api.ErrForbidden,
		},
		{
			name: "permission with wrong body param",
			uri:  "/path",
			body: "uid=invalid-type",
			endpoint: config.Endpoint{
				Method: "GET", Pattern: "/path", Scope: [][]string{{"A"}},
				Input: map[string]*config.Parameter{
					"uid": {ValidatorName: "uint", Kind: config.KindForm},
				},
				Fragments: []string{"path"},
			},
			method: "GET", path: "/path",
			permissions: []string{},
			err:         api.ErrForbidden,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			builder := &Builder{}

			// add active permissions
			builder.WithContext(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					a := api.Extract(r.Context()).Auth
					require.NotNil(t, a, "cannot access api.Auth form request context")

					a.Active = tc.permissions
					next.ServeHTTP(w, r)
				})
			})

			err := builder.Setup(strings.NewReader(baseConf))
			require.NoError(t, err)

			builder.conf.Endpoints = []*config.Endpoint{&tc.endpoint}

			err = builder.Bind(tc.method, tc.path, noOpHandler)
			require.NoError(t, err)

			handler, err := builder.Build(baseValidators)
			require.NoError(t, err)

			var (
				body     = strings.NewReader(tc.body)
				response = httptest.NewRecorder()
				request  = httptest.NewRequest(http.MethodGet, tc.uri, body)
			)

			// test request
			handler.ServeHTTP(response, request)
			require.NotNil(t, response.Body, "body")

			expectedStatus := api.GetErrorStatus(tc.err)
			require.Equalf(t, expectedStatus, response.Result().StatusCode, "http status")
		})
	}
}

func TestHandler_NilHandler(t *testing.T) {
	builder := &Builder{}

	err := builder.Setup(strings.NewReader(baseConf))
	require.NoError(t, err)

	builder.conf.Endpoints = []*config.Endpoint{
		{Name: "1", Method: "GET", Pattern: "/path", Fragments: []string{"path"}},
	}

	// nil handler
	builder.handlers = map[string]http.Handler{
		"1": nil,
	}

	handler, err := builder.Build(baseValidators)
	require.NoError(t, err)

	req, res := httptest.NewRequest("GET", "/path", nil), httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	require.Equal(t, api.ErrUncallableService.Status(), res.Result().StatusCode, "http status")
	expect := fmt.Sprintf("{\"status\":\"%s\"}", api.ErrUncallableService.Error())
	require.Equal(t, expect, res.Body.String(), "response body")
}

func TestHandlerRequestTooLarge(t *testing.T) {
	tt := []struct {
		name              string
		uriMax, uriSize   int
		bodyMax, bodySize int
		err               error
	}{
		{
			name:     "defaults -1",
			uriSize:  DefaultURILimit - 1,
			bodySize: DefaultBodyLimit - 1,
			err:      api.ErrUnknownService,
		},
		{
			name:     "defaults eq",
			uriSize:  DefaultURILimit,
			bodySize: DefaultBodyLimit,
			err:      api.ErrUnknownService,
		},
		{
			name:    "defaults uri",
			uriSize: DefaultURILimit + 1,
			err:     api.ErrURITooLong,
		},
		{
			name:     "defaults body",
			bodySize: DefaultBodyLimit + 1,
			err:      api.ErrBodyTooLarge,
		},
		{
			name:     "defaults both",
			uriSize:  DefaultURILimit + 1,
			bodySize: DefaultBodyLimit + 1,
			err:      api.ErrURITooLong,
		},

		{
			name:    "unlimited uri",
			uriMax:  -1,
			uriSize: DefaultURILimit + 1,
			err:     api.ErrUnknownService,
		},
		{
			name:     "unlimited body",
			bodyMax:  -1,
			bodySize: DefaultBodyLimit + 1,
			err:      api.ErrUnknownService,
		},
		{
			name:    "custom uri ok",
			uriMax:  50,
			uriSize: 50,
			err:     api.ErrUnknownService,
		},
		{
			name:    "custom uri",
			uriMax:  50,
			uriSize: 51,
			err:     api.ErrURITooLong,
		},
		{
			name:     "custom body ok",
			bodyMax:  50,
			bodySize: 50,
			err:      api.ErrUnknownService,
		},
		{
			name:     "custom body",
			bodyMax:  50,
			bodySize: 51,
			err:      api.ErrBodyTooLarge,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			b := &Builder{}

			err := b.Setup(strings.NewReader(baseConf))
			require.NoError(t, err)

			b.SetURILimit(tc.uriMax)
			b.SetBodyLimit(int64(tc.bodyMax))

			handler, err := b.Build(baseValidators)
			require.NoError(t, err)

			srv := httptest.NewServer(handler)
			defer srv.Close()

			host := fmt.Sprintf("%s/", srv.URL)

			// build fake uri and body according to test sizes
			var (
				fakeURI  = strings.Repeat("a", tc.uriSize)
				fakeBody = strings.Repeat("a", tc.bodySize)
			)
			// remove 1 to take the '/' into account
			if len(fakeURI) > 0 {
				fakeURI = strings.TrimSuffix(fakeURI, "a")
			}

			req, err := http.NewRequest("POST", host+fakeURI, strings.NewReader(fakeBody))
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			var expect int = http.StatusOK
			if tc.err != nil {
				cast, ok := tc.err.(api.Err)
				require.True(t, ok, "invalid error type")
				expect = cast.Status()
			}
			require.Equal(t, expect, res.StatusCode, "http status")
		})
	}
}

func TestHandler_EndpointErrors(t *testing.T) {
	noOpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runtime.Respond(w, nil, nil)
	})

	tt := []struct {
		name     string
		endpoint config.Endpoint
		// handler
		bindMethod, bindPath string
		// request
		method, url string
		contentType string
		body        string
		permissions []string
		err         error
		errReason   string
	}{
		// service match
		{
			name: "unknown service method",
			endpoint: config.Endpoint{
				Method: "GET", Pattern: "/path", Fragments: []string{"path"},
			},
			bindMethod: "GET", bindPath: "/path",
			method: "POST", url: "/",
			body:        ``,
			permissions: []string{},
			err:         api.ErrUnknownService,
		},
		{
			name: "unknown service path",
			endpoint: config.Endpoint{
				Method: "GET", Pattern: "/", Fragments: []string{},
			},
			bindMethod: "GET", bindPath: "/",
			method: "GET", url: "/invalid",
			body:        ``,
			permissions: []string{},
			err:         api.ErrUnknownService,
		},
		{
			name: "nominal",
			endpoint: config.Endpoint{
				Method: "GET", Pattern: "/", Fragments: []string{},
			},
			bindMethod: "GET", bindPath: "/",
			method: "GET", url: "/",
			body:        ``,
			permissions: []string{},
			err:         nil,
		},

		// invalid uri param -> unknown service
		{
			name: "invalid uri param",
			endpoint: config.Endpoint{
				Method: "GET", Pattern: "/a/{id}/b", Fragments: []string{"a", "{id}", "b"},
				Input: map[string]*config.Parameter{
					"{id}": {ValidatorName: "uint", Kind: config.KindURI, Rename: "ID"},
				},
				Captures: []*config.BraceCapture{
					{Index: 1, Name: "id"},
				},
			},
			bindMethod: "GET", bindPath: "/a/{id}/b",
			method: "GET", url: "/a/invalid/b",
			body:        ``,
			permissions: []string{},
			err:         api.ErrUnknownService,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			builder := &Builder{}
			err := builder.Setup(strings.NewReader(baseConf))
			require.NoError(t, err, "setup")

			builder.conf.Endpoints = []*config.Endpoint{&tc.endpoint}

			err = builder.Bind(tc.bindMethod, tc.bindPath, noOpHandler)
			require.NoError(t, err)

			handler, err := builder.Build(baseValidators)
			require.NoError(t, err, "build")

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
			require.Equalf(t, api.GetErrorStatus(tc.err), response.Result().StatusCode, "http status")

			if len(tc.errReason) < 1 {
				return
			}

			type JSONError struct {
				Status string `json:"status"`
			}
			var parsedError JSONError
			err = json.NewDecoder(response.Body).Decode(&parsedError)
			require.NoError(t, err, "parse body")
			require.Equal(t, tc.errReason, parsedError.Status, "error reason")
		})
	}
}
