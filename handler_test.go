package aicra

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.xdrm.io/go/aicra/api"
)

func TestWith(t *testing.T) {
	builder := &Builder{}
	if err := addBuiltinTypes(builder); err != nil {
		t.Fatalf("unexpected error <%v>", err)
	}

	// build @n middlewares that take data from context and increment it
	n := 1024

	type ckey int
	const key ckey = 0

	middleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			newr := r

			// first time -> store 1
			value := r.Context().Value(key)
			if value == nil {
				newr = r.WithContext(context.WithValue(r.Context(), key, int(1)))
				next(w, newr)
				return
			}

			// get value and increment
			cast, ok := value.(int)
			if !ok {
				t.Fatalf("value is not an int")
			}
			cast++
			newr = r.WithContext(context.WithValue(r.Context(), key, cast))
			next(w, newr)
		}
	}

	// add middleware @n times
	for i := 0; i < n; i++ {
		builder.With(middleware)
	}

	config := strings.NewReader(`[ { "method": "GET", "path": "/path", "scope": [[]], "info": "info", "in": {}, "out": {} } ]`)
	err := builder.Setup(config)
	if err != nil {
		t.Fatalf("setup: unexpected error <%v>", err)
	}

	pathHandler := func(ctx api.Ctx) (*struct{}, api.Err) {
		// write value from middlewares into response
		value := ctx.Req.Context().Value(key)
		if value == nil {
			t.Fatalf("nothing found in context")
		}
		cast, ok := value.(int)
		if !ok {
			t.Fatalf("cannot cast context data to int")
		}
		// write to response
		ctx.Res.Write([]byte(fmt.Sprintf("#%d#", cast)))

		return nil, api.ErrSuccess
	}

	if err := builder.Bind(http.MethodGet, "/path", pathHandler); err != nil {
		t.Fatalf("bind: unexpected error <%v>", err)
	}

	handler, err := builder.Build()
	if err != nil {
		t.Fatalf("build: unexpected error <%v>", err)
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/path", &bytes.Buffer{})

	// test request
	handler.ServeHTTP(response, request)
	if response.Body == nil {
		t.Fatalf("response has no body")
	}
	token := fmt.Sprintf("#%d#", n)
	if !strings.Contains(response.Body.String(), token) {
		t.Fatalf("expected '%s' to be in response <%s>", token, response.Body.String())
	}

}
