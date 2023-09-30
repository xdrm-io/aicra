package aicra_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xdrm-io/aicra"
	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/aicra/runtime"
)

const staticConfig = `[
	{
		"method": "GET",
		"path": "/users/123",
		"scope": [],
		"info": "info",
		"in": {},
		"out": {}
	}
]`
const staticOutConfig = `[
	{
		"method": "GET",
		"path": "/users/123",
		"scope": [],
		"info": "info",
		"in": {},
		"out": {
			"id": {"info":"info","name":"ID","type":"int"}
		}
	}
]`
const uriConfig = `[
	{
		"method": "GET",
		"path": "/users/{id}",
		"scope": [],
		"info": "info",
		"in": {
			"{id}": {"info":"info","name":"ID","type":"int"}
		},
		"out": {}
	}
]`
const getConfig = `[
	{
		"method": "GET",
		"path": "/users",
		"scope": [],
		"info": "info",
		"in": {
			"GET@id": {"info":"info","name":"ID","type":"int"}
		},
		"out": {}
	}
]`
const formConfig = `[
	{
		"method": "GET",
		"path": "/users",
		"scope": [],
		"info": "info",
		"in": {
			"id": {"info":"info","name":"ID","type":"int"}
		},
		"out": {}
	}
]`
const staticMultiConfig = `[
	{
		"method": "GET",
		"path": "/users/123",
		"scope": [],
		"info": "info",
		"in": {},
		"out": {}
	},
	{
		"method": "POST",
		"path": "/users/123",
		"scope": [],
		"info": "info",
		"in": {},
		"out": {}
	},
	{
		"method": "GET",
		"path": "/users/456",
		"scope": [],
		"info": "info",
		"in": {},
		"out": {}
	},
	{
		"method": "POST",
		"path": "/users/456",
		"scope": [],
		"info": "info",
		"in": {},
		"out": {}
	}
]`
const uriMultiConfig = `[
	{
		"method": "GET",
		"path": "/users/{id}",
		"scope": [],
		"info": "info",
		"in": {
			"{id}": {"info":"info","name":"ID","type":"int"}
		},
		"out": {}
	},
	{
		"method": "POST",
		"path": "/users/{id}",
		"scope": [],
		"info": "info",
		"in": {
			"{id}": {"info":"info","name":"ID","type":"int"}
		},
		"out": {}
	},
	{
		"method": "GET",
		"path": "/articles/{id}",
		"scope": [],
		"info": "info",
		"in": {
			"{id}": {"info":"info","name":"ID","type":"int"}
		},
		"out": {}
	},
	{
		"method": "POST",
		"path": "/articles/{id}",
		"scope": [],
		"info": "info",
		"in": {
			"{id}": {"info":"info","name":"ID","type":"int"}
		},
		"out": {}
	}
]`
const getMultiConfig = `[
	{
		"method": "GET",
		"path": "/users",
		"scope": [],
		"info": "info",
		"in": {
			"GET@id": {"info":"info","name":"ID","type":"int"}
		},
		"out": {}
	},
	{
		"method": "POST",
		"path": "/users",
		"scope": [],
		"info": "info",
		"in": {
			"GET@id": {"info":"info","name":"ID","type":"int"}
		},
		"out": {}
	},
	{
		"method": "GET",
		"path": "/articles",
		"scope": [],
		"info": "info",
		"in": {
			"GET@id": {"info":"info","name":"ID","type":"int"}
		},
		"out": {}
	},
	{
		"method": "POST",
		"path": "/articles",
		"scope": [],
		"info": "info",
		"in": {
			"GET@id": {"info":"info","name":"ID","type":"int"}
		},
		"out": {}
	}
]`
const formMultiConfig = `[
	{
		"method": "GET",
		"path": "/users",
		"scope": [],
		"info": "info",
		"in": {
			"id": {"info":"info","name":"ID","type":"int"}
		},
		"out": {}
	},
	{
		"method": "POST",
		"path": "/users",
		"scope": [],
		"info": "info",
		"in": {
			"id": {"info":"info","name":"ID","type":"int"}
		},
		"out": {}
	},
	{
		"method": "GET",
		"path": "/articles",
		"scope": [],
		"info": "info",
		"in": {
			"id": {"info":"info","name":"ID","type":"int"}
		},
		"out": {}
	},
	{
		"method": "POST",
		"path": "/articles",
		"scope": [],
		"info": "info",
		"in": {
			"id": {"info":"info","name":"ID","type":"int"}
		},
		"out": {}
	}
]`

func noOpHandler(w http.ResponseWriter, r *http.Request) {}
func outHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"id": 123,
	}
	runtime.Respond(w, data, nil)
}

func Benchmark1StaticRouteMatch(b *testing.B) {
	builder := &aicra.Builder{}

	err := builder.Setup(strings.NewReader(staticConfig))
	if err != nil {
		b.Fatalf("cannot setup: %s", err)
	}
	err = builder.Bind("GET", "/users/123", noOpHandler)
	if err != nil {
		b.Fatalf("cannot bind: %s", err)
	}
	srv, err := builder.Build(config.Validators{})
	if err != nil {
		b.Fatalf("cannot build: %s", err)
	}

	req, _ := http.NewRequest("GET", "/users/123", nil)
	res := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(res, req)
	}
}
func Benchmark1StaticOutRouteMatch(b *testing.B) {
	builder := &aicra.Builder{}

	err := builder.Setup(strings.NewReader(staticOutConfig))
	if err != nil {
		b.Fatalf("cannot setup: %s", err)
	}
	err = builder.Bind("GET", "/users/123", outHandler)
	if err != nil {
		b.Fatalf("cannot bind: %s", err)
	}
	srv, err := builder.Build(config.Validators{})
	if err != nil {
		b.Fatalf("cannot build: %s", err)
	}

	req, _ := http.NewRequest("GET", "/users/123", nil)
	res := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(res, req)
	}
}

func Benchmark1OverNStaticRouteMatch(b *testing.B) {
	builder := &aicra.Builder{}

	err := builder.Setup(strings.NewReader(staticMultiConfig))
	if err != nil {
		b.Fatalf("cannot setup: %s", err)
	}

	routes := [][2]string{
		{"GET", "/users/123"},
		{"POST", "/users/123"},
		{"GET", "/users/456"},
		{"POST", "/users/456"},
	}
	for _, route := range routes {
		if err := builder.Bind(route[0], route[1], noOpHandler); err != nil {
			b.Fatalf("cannot bind: %s", err)
		}

	}
	srv, err := builder.Build(config.Validators{})
	if err != nil {
		b.Fatalf("cannot build: %s", err)
	}

	req, _ := http.NewRequest("GET", "/users/123", nil)
	res := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(res, req)
	}
}

func Benchmark1UriRouteMatch(b *testing.B) {
	builder := &aicra.Builder{}

	err := builder.Setup(strings.NewReader(uriConfig))
	if err != nil {
		b.Fatalf("cannot setup: %s", err)
	}
	err = builder.Bind("GET", "/users/{id}", noOpHandler)
	if err != nil {
		b.Fatalf("cannot bind: %s", err)
	}
	srv, err := builder.Build(config.Validators{})
	if err != nil {
		b.Fatalf("cannot build: %s", err)
	}

	req, _ := http.NewRequest("GET", "/users/123", nil)
	res := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(res, req)
	}
}

func Benchmark1OverNUriRouteMatch(b *testing.B) {
	builder := &aicra.Builder{}

	err := builder.Setup(strings.NewReader(uriMultiConfig))
	if err != nil {
		b.Fatalf("cannot setup: %s", err)
	}

	routes := [][2]string{
		{"GET", "/users/{id}"},
		{"POST", "/users/{id}"},
		{"GET", "/articles/{id}"},
		{"POST", "/articles/{id}"},
	}
	for _, route := range routes {
		if err := builder.Bind(route[0], route[1], noOpHandler); err != nil {
			b.Fatalf("cannot bind: %s", err)
		}

	}
	srv, err := builder.Build(config.Validators{})
	if err != nil {
		b.Fatalf("cannot build: %s", err)
	}

	req, _ := http.NewRequest("GET", "/users/123", nil)
	res := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(res, req)
	}
}

func Benchmark1GetRouteMatch(b *testing.B) {
	builder := &aicra.Builder{}

	err := builder.Setup(strings.NewReader(getConfig))
	if err != nil {
		b.Fatalf("cannot setup: %s", err)
	}
	err = builder.Bind("GET", "/users", noOpHandler)
	if err != nil {
		b.Fatalf("cannot bind: %s", err)
	}
	srv, err := builder.Build(config.Validators{})
	if err != nil {
		b.Fatalf("cannot build: %s", err)
	}

	req, _ := http.NewRequest("GET", "/users?id=123", nil)
	res := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(res, req)
	}
}
func Benchmark1OverNGetRouteMatch(b *testing.B) {
	builder := &aicra.Builder{}

	err := builder.Setup(strings.NewReader(getMultiConfig))
	if err != nil {
		b.Fatalf("cannot setup: %s", err)
	}

	routes := [][2]string{
		{"GET", "/users"},
		{"POST", "/users"},
		{"GET", "/articles"},
		{"POST", "/articles"},
	}
	for _, route := range routes {
		if err := builder.Bind(route[0], route[1], noOpHandler); err != nil {
			b.Fatalf("cannot bind: %s", err)
		}

	}
	srv, err := builder.Build(config.Validators{})
	if err != nil {
		b.Fatalf("cannot build: %s", err)
	}

	req, _ := http.NewRequest("GET", "/users?id=123", nil)
	res := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(res, req)
	}
}

func Benchmark1URLEncodedRouteMatch(b *testing.B) {
	builder := &aicra.Builder{}

	err := builder.Setup(strings.NewReader(formConfig))
	if err != nil {
		b.Fatalf("cannot setup: %s", err)
	}
	err = builder.Bind("GET", "/users", noOpHandler)
	if err != nil {
		b.Fatalf("cannot bind: %s", err)
	}
	srv, err := builder.Build(config.Validators{})
	if err != nil {
		b.Fatalf("cannot build: %s", err)
	}

	req, _ := http.NewRequest("GET", "/users", strings.NewReader("id=123"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(res, req)
	}
}
func Benchmark1OverNURLEncodedRouteMatch(b *testing.B) {
	builder := &aicra.Builder{}

	err := builder.Setup(strings.NewReader(formMultiConfig))
	if err != nil {
		b.Fatalf("cannot setup: %s", err)
	}

	routes := [][2]string{
		{"GET", "/users"},
		{"POST", "/users"},
		{"GET", "/articles"},
		{"POST", "/articles"},
	}
	for _, route := range routes {
		if err := builder.Bind(route[0], route[1], noOpHandler); err != nil {
			b.Fatalf("cannot bind: %s", err)
		}

	}
	srv, err := builder.Build(config.Validators{})
	if err != nil {
		b.Fatalf("cannot build: %s", err)
	}

	req, _ := http.NewRequest("GET", "/users", strings.NewReader("id=123"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(res, req)
	}
}

func Benchmark1JsonRouteMatch(b *testing.B) {
	builder := &aicra.Builder{}

	err := builder.Setup(strings.NewReader(formConfig))
	if err != nil {
		b.Fatalf("cannot setup: %s", err)
	}
	err = builder.Bind("GET", "/users", noOpHandler)
	if err != nil {
		b.Fatalf("cannot bind: %s", err)
	}
	srv, err := builder.Build(config.Validators{})
	if err != nil {
		b.Fatalf("cannot build: %s", err)
	}

	req, _ := http.NewRequest("GET", "/users", strings.NewReader(`{"id":123}`))
	req.Header.Add("Content-Type", "application/json")
	res := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(res, req)
	}
}
func Benchmark1OverNJsonRouteMatch(b *testing.B) {
	builder := &aicra.Builder{}

	err := builder.Setup(strings.NewReader(formMultiConfig))
	if err != nil {
		b.Fatalf("cannot setup: %s", err)
	}

	routes := [][2]string{
		{"GET", "/users"},
		{"POST", "/users"},
		{"GET", "/articles"},
		{"POST", "/articles"},
	}
	for _, route := range routes {
		if err := builder.Bind(route[0], route[1], noOpHandler); err != nil {
			b.Fatalf("cannot bind: %s", err)
		}

	}
	srv, err := builder.Build(config.Validators{})
	if err != nil {
		b.Fatalf("cannot build: %s", err)
	}

	req, _ := http.NewRequest("GET", "/users", strings.NewReader(`{"id":123}`))
	req.Header.Add("Content-Type", "application/json")
	res := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(res, req)
	}
}

func Benchmark1MultipartRouteMatch(b *testing.B) {
	builder := &aicra.Builder{}

	err := builder.Setup(strings.NewReader(formConfig))
	if err != nil {
		b.Fatalf("cannot setup: %s", err)
	}
	err = builder.Bind("GET", "/users", noOpHandler)
	if err != nil {
		b.Fatalf("cannot bind: %s", err)
	}
	srv, err := builder.Build(config.Validators{})
	if err != nil {
		b.Fatalf("cannot build: %s", err)
	}

	req, _ := http.NewRequest("GET", "/users", strings.NewReader(`--x
	Content-Disposition: form-data; name="id"

	123
	--x--`))
	req.Header.Add("Content-Type", "multipart/form-data; boundary=x")
	res := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(res, req)
	}
}
func Benchmark1OverNMultipartRouteMatch(b *testing.B) {
	builder := &aicra.Builder{}

	err := builder.Setup(strings.NewReader(formMultiConfig))
	if err != nil {
		b.Fatalf("cannot setup: %s", err)
	}

	routes := [][2]string{
		{"GET", "/users"},
		{"POST", "/users"},
		{"GET", "/articles"},
		{"POST", "/articles"},
	}
	for _, route := range routes {
		if err := builder.Bind(route[0], route[1], noOpHandler); err != nil {
			b.Fatalf("cannot bind: %s", err)
		}

	}
	srv, err := builder.Build(config.Validators{})
	if err != nil {
		b.Fatalf("cannot build: %s", err)
	}

	req, _ := http.NewRequest("GET", "/users", strings.NewReader(`--x
	Content-Disposition: form-data; name="id"

	123
	--x--`))
	req.Header.Add("Content-Type", "multipart/form-data; boundary=x")
	res := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(res, req)
	}
}
