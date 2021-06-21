package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/internal/ctx"
)

func TestContextGetRequest(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/random", nil)
	if err != nil {
		t.Fatalf("cannot create http request: %s", err)
	}

	// store in bare context
	c := context.Background()
	c = context.WithValue(c, ctx.Request, req)

	// fetch from context
	fetched := api.GetRequest(c)
	if fetched != req {
		t.Fatalf("fetched http request %v ; expected %v", fetched, req)
	}
}
func TestContextGetNilRequest(t *testing.T) {
	// fetch from bare context
	fetched := api.GetRequest(context.Background())
	if fetched != nil {
		t.Fatalf("fetched http request %v from empty context; expected nil", fetched)
	}
}

func TestContextGetResponseWriter(t *testing.T) {
	res := httptest.NewRecorder()

	// store in bare context
	c := context.Background()
	c = context.WithValue(c, ctx.Response, res)

	// fetch from context
	fetched := api.GetResponseWriter(c)
	if fetched != res {
		t.Fatalf("fetched http response writer %v ; expected %v", fetched, res)
	}
}

func TestContextGetNilResponseWriter(t *testing.T) {
	// fetch from bare context
	fetched := api.GetResponseWriter(context.Background())
	if fetched != nil {
		t.Fatalf("fetched http response writer %v from empty context; expected nil", fetched)
	}
}

func TestContextGetAuth(t *testing.T) {
	auth := &api.Auth{}

	// store in bare context
	c := context.Background()
	c = context.WithValue(c, ctx.Auth, auth)

	// fetch from context
	fetched := api.GetAuth(c)
	if fetched != auth {
		t.Fatalf("fetched api auth %v ; expected %v", fetched, auth)
	}
}

func TestContextGetNilAuth(t *testing.T) {
	// fetch from bare context
	fetched := api.GetAuth(context.Background())
	if fetched != nil {
		t.Fatalf("fetched api auth %v from empty context; expected nil", fetched)
	}
}
