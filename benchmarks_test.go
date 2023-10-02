package aicra

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/internal/config"
)

func noOpHandler(w http.ResponseWriter, r *http.Request) {}

func createBuilder(b *testing.B) *Builder {
	b.Helper()

	builder := &Builder{}
	err := builder.Setup(strings.NewReader(baseConf))
	require.NoError(b, err)
	return builder
}

type route struct {
	method string
	path   string
}

func createRoutes(b *testing.B, n uint) []route {
	b.Helper()

	routes := make([]route, n)
	for i := uint(0); i < n; i += 4 {
		uri := "/users/n" + strconv.FormatUint(uint64(i), 10)
		routes[i+0] = route{"GET", uri}
		routes[i+1] = route{"POST", uri}
		routes[i+2] = route{"PUT", uri}
		routes[i+3] = route{"DELETE", uri}
	}
	return routes
}

// NRoutes defines the number of endpoints on which to run the benchmarks
const NRoutes = 100

func static(b *testing.B) (http.Handler, []route) {
	b.Helper()

	var (
		builder = createBuilder(b)
		routes  = createRoutes(b, NRoutes)
	)
	for i, route := range routes {
		builder.conf.Endpoints = append(builder.conf.Endpoints, &config.Endpoint{
			Name:      strconv.Itoa(i),
			Method:    route.method,
			Pattern:   route.path,
			Fragments: config.URIFragments(route.path),
		})

		err := builder.Bind(route.method, route.path, noOpHandler)
		require.NoError(b, err)

	}
	srv, err := builder.Build(baseValidators)
	require.NoError(b, err)

	return srv, routes
}

func BenchmarkStaticFirst(b *testing.B) {
	var (
		handler, routes = static(b)
		first           = routes[0]
		req, _          = http.NewRequest(first.method, first.path, nil)
		res             = httptest.NewRecorder()
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}
func BenchmarkStaticLast(b *testing.B) {
	var (
		handler, routes = static(b)
		last            = routes[len(routes)-1]
		req, _          = http.NewRequest(last.method, last.path, nil)
		res             = httptest.NewRecorder()
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}

func uri(b *testing.B) (http.Handler, []route) {
	builder := createBuilder(b)

	routes := createRoutes(b, NRoutes)
	for i, route := range routes {
		path := route.path + "/{id}"
		fragments := config.URIFragments(path)
		builder.conf.Endpoints = append(builder.conf.Endpoints, &config.Endpoint{
			Name:      strconv.Itoa(i),
			Method:    route.method,
			Pattern:   path,
			Fragments: fragments,
			Input: map[string]*config.Parameter{
				"{id}": {Rename: "ID", ValidatorName: "uint"},
			},
			Captures: []*config.BraceCapture{
				{Index: len(fragments) - 1, Name: "id"},
			},
		})
		err := builder.Bind(route.method, path, noOpHandler)
		require.NoError(b, err)
	}
	srv, err := builder.Build(baseValidators)
	require.NoError(b, err)
	return srv, routes
}

func BenchmarkURIFirst(b *testing.B) {
	var (
		handler, routes = uri(b)
		first           = routes[0]
		req, _          = http.NewRequest(first.method, first.path+"/123", nil)
		res             = httptest.NewRecorder()
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}

func BenchmarkURILast(b *testing.B) {
	var (
		handler, routes = uri(b)
		last            = routes[len(routes)-1]
		req, _          = http.NewRequest(last.method, last.path+"/123", nil)
		res             = httptest.NewRecorder()
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}

func query(b *testing.B) (http.Handler, []route) {
	builder := createBuilder(b)

	routes := createRoutes(b, NRoutes)
	for i, route := range routes {
		builder.conf.Endpoints = append(builder.conf.Endpoints, &config.Endpoint{
			Name:      strconv.Itoa(i),
			Method:    route.method,
			Pattern:   route.path,
			Fragments: config.URIFragments(route.path),
			Input: map[string]*config.Parameter{
				"?id": {Rename: "ID", ValidatorName: "uint"},
			},
		})
		err := builder.Bind(route.method, route.path, noOpHandler)
		require.NoError(b, err)
	}
	srv, err := builder.Build(baseValidators)
	require.NoError(b, err)
	return srv, routes
}

func BenchmarkQueryFirst(b *testing.B) {
	var (
		handler, routes = query(b)
		first           = routes[0]
		req, _          = http.NewRequest(first.method, first.path+"?id=123", nil)
		res             = httptest.NewRecorder()
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}
func BenchmarkQueryLast(b *testing.B) {
	var (
		handler, routes = query(b)
		first           = routes[0]
		req, _          = http.NewRequest(first.method, first.path+"?id=123", nil)
		res             = httptest.NewRecorder()
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}

func form(b *testing.B) (http.Handler, []route) {
	builder := createBuilder(b)

	routes := createRoutes(b, NRoutes)
	for i, route := range routes {
		builder.conf.Endpoints = append(builder.conf.Endpoints, &config.Endpoint{
			Name:      strconv.Itoa(i),
			Method:    route.method,
			Pattern:   route.path,
			Fragments: config.URIFragments(route.path),
			Input: map[string]*config.Parameter{
				"id": {Rename: "ID", ValidatorName: "uint"},
			},
		})
		err := builder.Bind(route.method, route.path, noOpHandler)
		require.NoError(b, err)
	}
	srv, err := builder.Build(baseValidators)
	require.NoError(b, err)
	return srv, routes
}

func BenchmarkURLEncodedFirst(b *testing.B) {
	var (
		handler, routes = form(b)
		first           = routes[0]
		req, _          = http.NewRequest(first.method, first.path, strings.NewReader("id=123"))
		res             = httptest.NewRecorder()
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}
func BenchmarkURLEncodedLast(b *testing.B) {
	var (
		handler, routes = form(b)
		last            = routes[len(routes)-1]
		req, _          = http.NewRequest(last.method, last.path, strings.NewReader("id=123"))
		res             = httptest.NewRecorder()
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}

func BenchmarkJSONFirst(b *testing.B) {
	var (
		handler, routes = form(b)
		first           = routes[0]
		req, _          = http.NewRequest(first.method, first.path, strings.NewReader(`{"id":123}`))
		res             = httptest.NewRecorder()
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}
func BenchmarkJSONLast(b *testing.B) {
	var (
		handler, routes = form(b)
		last            = routes[len(routes)-1]
		req, _          = http.NewRequest(last.method, last.path, strings.NewReader(`{"id":123}`))
		res             = httptest.NewRecorder()
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}

func BenchmarkMultipartFirst(b *testing.B) {
	var (
		handler, routes = form(b)
		first           = routes[0]
		req, _          = http.NewRequest(first.method, first.path, strings.NewReader(`--x
		Content-Disposition: form-data; name="id"

		123
		--x--`))
		res = httptest.NewRecorder()
	)

	req.Header.Add("Content-Type", "multipart/form-data; boundary=x")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}
func BenchmarkMultipartLast(b *testing.B) {
	var (
		handler, routes = form(b)
		last            = routes[len(routes)-1]
		req, _          = http.NewRequest(last.method, last.path, strings.NewReader(`--x
		Content-Disposition: form-data; name="id"

		123
		--x--`))
		res = httptest.NewRecorder()
	)

	req.Header.Add("Content-Type", "multipart/form-data; boundary=x")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}
