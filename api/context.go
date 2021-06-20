package api

import (
	"context"
	"net/http"

	"github.com/xdrm-io/aicra/internal/ctx"
)

// GetRequest extracts the current request from a context.Context
func GetRequest(c context.Context) *http.Request {
	var (
		raw      = c.Value(ctx.Request)
		cast, ok = raw.(*http.Request)
	)
	if !ok {
		return nil
	}
	return cast
}

// GetResponseWriter extracts the response writer from a context.Context
func GetResponseWriter(c context.Context) http.ResponseWriter {
	var (
		raw      = c.Value(ctx.Response)
		cast, ok = raw.(http.ResponseWriter)
	)
	if !ok {
		return nil
	}
	return cast
}

// GetAuth returns the api.Auth associated with this request from a context.Context
func GetAuth(c context.Context) *Auth {
	var (
		raw      = c.Value(ctx.Auth)
		cast, ok = raw.(*Auth)
	)
	if !ok {
		return nil
	}
	return cast
}
