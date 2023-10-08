package aicra

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/aicra/runtime"
)

func noOpHandler(w http.ResponseWriter, r *http.Request) {}

func formHandler(b *testing.B) http.HandlerFunc {
	b.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := runtime.ParseForm(r)
		require.NoError(b, err)
	}
}

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
	methods := []string{"GET", "POST", "PUT", "DELETE"}

	routes := make([]route, n)
	methodi := uint8(0)
	for i := uint(0); i < n; i++ {
		uri := "/users/n" + strconv.FormatUint(uint64(i), 10)
		routes[i] = route{methods[methodi], uri}
		methodi = (methodi + 1) % uint8(len(methods))
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

func uriMulti(b *testing.B, nVars int) (http.Handler, []route) {
	builder := createBuilder(b)
	builder.SetURILimit(1e6)

	routes := createRoutes(b, NRoutes)
	var vars strings.Builder
	for i, route := range routes {
		vars.Reset()
		var (
			input    = make(map[string]*config.Parameter, nVars)
			captures = make([]*config.BraceCapture, nVars)
		)
		for v := 0; v < nVars; v++ {
			vars.WriteString(`/{a` + strconv.Itoa(v) + `}`)
			input[`{a`+strconv.Itoa(v)+`}`] = &config.Parameter{Rename: "A" + strconv.Itoa(v), ValidatorName: "uint"}
			captures[v] = &config.BraceCapture{Index: 2 + v, Name: "a" + strconv.Itoa(v)}
		}

		path := route.path + vars.String()
		fragments := config.URIFragments(path)
		builder.conf.Endpoints = append(builder.conf.Endpoints, &config.Endpoint{
			Name:      strconv.Itoa(i),
			Method:    route.method,
			Pattern:   path,
			Fragments: fragments,
			Input:     input,
			Captures:  captures,
		})
		err := builder.Bind(route.method, path, noOpHandler)
		require.NoError(b, err)
	}
	srv, err := builder.Build(baseValidators)
	require.NoError(b, err)
	return srv, routes
}

func BenchmarkURI100ParamsFirst(b *testing.B) {
	var (
		handler, routes = uriMulti(b, 100)
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
func BenchmarkURI100ParamsLast(b *testing.B) {
	var (
		handler, routes = uriMulti(b, 100)
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

func queryMulti(b *testing.B, nVars int) (http.Handler, []route) {
	builder := createBuilder(b)
	builder.SetURILimit(1e6)

	routes := createRoutes(b, NRoutes)
	var vars strings.Builder
	for i, route := range routes {
		vars.Reset()
		var (
			path  = route.path
			input = make(map[string]*config.Parameter, nVars)
		)
		for v := 0; v < nVars; v++ {
			if v == 0 {
				vars.WriteString(`?a` + strconv.Itoa(v) + `=` + strconv.Itoa(v))
			} else {
				vars.WriteString(`&a` + strconv.Itoa(v) + `=` + strconv.Itoa(v))
			}
			input[`?a`+strconv.Itoa(v)] = &config.Parameter{Rename: "A" + strconv.Itoa(v), ValidatorName: "uint", Kind: config.KindQuery}
		}
		uri := path + vars.String()

		fragments := config.URIFragments(path)
		builder.conf.Endpoints = append(builder.conf.Endpoints, &config.Endpoint{
			Name:      strconv.Itoa(i),
			Method:    route.method,
			Pattern:   path,
			Fragments: fragments,
			Input:     input,
		})
		err := builder.Bind(route.method, path, noOpHandler)
		require.NoError(b, err)

		// return the full uri
		routes[i].path = uri
	}
	srv, err := builder.Build(baseValidators)
	require.NoError(b, err)
	return srv, routes
}

func BenchmarkQuery100ParamsFirst(b *testing.B) {
	var (
		handler, routes = queryMulti(b, 100)
		first           = routes[0]
		req, _          = http.NewRequest(first.method, first.path, strings.NewReader(`--x
		Content-Disposition: form-data; name="id"

		123
		--x--`))
		res = httptest.NewRecorder()
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}
func BenchmarkQuery100ParamsLast(b *testing.B) {
	var (
		handler, routes = queryMulti(b, 100)
		last            = routes[len(routes)-1]
		req, _          = http.NewRequest(last.method, last.path, strings.NewReader(`--x
		Content-Disposition: form-data; name="id"

		123
		--x--`))
		res = httptest.NewRecorder()
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}

func urlencodedMulti(b *testing.B, nVars int) (http.Handler, []route, []byte) {
	builder := createBuilder(b)

	routes := createRoutes(b, NRoutes)
	var vars strings.Builder
	for i, route := range routes {
		vars.Reset()
		input := make(map[string]*config.Parameter, nVars)
		for v := 0; v < nVars; v++ {
			input[`a`+strconv.Itoa(v)] = &config.Parameter{Rename: "A" + strconv.Itoa(v), ValidatorName: "uint", Kind: config.KindForm}
		}

		builder.conf.Endpoints = append(builder.conf.Endpoints, &config.Endpoint{
			Name:      strconv.Itoa(i),
			Method:    route.method,
			Pattern:   route.path,
			Fragments: config.URIFragments(route.path),
			Input:     input,
		})
		err := builder.Bind(route.method, route.path, formHandler(b))
		require.NoError(b, err)
	}

	// build body
	query := make(url.Values, nVars)
	for v := 0; v < nVars; v++ {
		query.Add(`a`+strconv.Itoa(v), strconv.Itoa(v))
	}
	body := []byte(query.Encode())

	srv, err := builder.Build(baseValidators)
	require.NoError(b, err)
	return srv, routes, body
}

func BenchmarkURLEncoded100ParamsFirst(b *testing.B) {
	var (
		handler, routes, body = urlencodedMulti(b, 100)
		first                 = routes[0]
		req, _                = http.NewRequest(first.method, first.path, bytes.NewReader(body))
		res                   = httptest.NewRecorder()
	)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}
func BenchmarkURLEncoded100ParamsLast(b *testing.B) {
	var (
		handler, routes, body = urlencodedMulti(b, 100)
		last                  = routes[len(routes)-1]
		req, _                = http.NewRequest(last.method, last.path, bytes.NewReader(body))
		res                   = httptest.NewRecorder()
	)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}

func jsonMulti(b *testing.B, nVars int) (http.Handler, []route, []byte) {
	builder := createBuilder(b)

	routes := createRoutes(b, NRoutes)
	var vars strings.Builder
	for i, route := range routes {
		vars.Reset()
		input := make(map[string]*config.Parameter, nVars)
		for v := 0; v < nVars; v++ {
			input[`a`+strconv.Itoa(v)] = &config.Parameter{Rename: "A" + strconv.Itoa(v), ValidatorName: "uint", Kind: config.KindForm}
		}

		builder.conf.Endpoints = append(builder.conf.Endpoints, &config.Endpoint{
			Name:      strconv.Itoa(i),
			Method:    route.method,
			Pattern:   route.path,
			Fragments: config.URIFragments(route.path),
			Input:     input,
		})
		err := builder.Bind(route.method, route.path, formHandler(b))
		require.NoError(b, err)
	}

	// build body
	form := make(map[string]int, nVars)
	for i, l := 0, len(form); i < l; i++ {
		form[`a`+strconv.Itoa(i)] = i
	}
	body, err := json.Marshal(form)
	require.NoError(b, err)

	srv, err := builder.Build(baseValidators)
	require.NoError(b, err)
	return srv, routes, body
}

func BenchmarkJSON100ParamsFirst(b *testing.B) {
	var (
		handler, routes, body = jsonMulti(b, 100)
		first                 = routes[0]
		req, _                = http.NewRequest(first.method, first.path, bytes.NewReader(body))
		res                   = httptest.NewRecorder()
	)
	req.Header.Add("Content-Type", "application/json")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}
func BenchmarkJSON100ParamsLast(b *testing.B) {
	var (
		handler, routes, body = jsonMulti(b, 100)
		last                  = routes[len(routes)-1]
		req, _                = http.NewRequest(last.method, last.path, bytes.NewReader(body))
		res                   = httptest.NewRecorder()
	)
	req.Header.Add("Content-Type", "application/json")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(res, req)
	}
}
