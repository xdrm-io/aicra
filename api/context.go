package api

import (
	"context"
	"net/http"

	"git.xdrm.io/go/aicra/internal/ctx"
)

// Context is a simple wrapper around context.Context that adds helper methods
// to access additional information
type Context struct{ context.Context }

// Request current request
func (c Context) Request() *http.Request {
	var (
		raw      = c.Value(ctx.Request)
		cast, ok = raw.(*http.Request)
	)
	if !ok {
		return nil
	}
	return cast
}

// ResponseWriter for this request
func (c Context) ResponseWriter() http.ResponseWriter {
	var (
		raw      = c.Value(ctx.Response)
		cast, ok = raw.(http.ResponseWriter)
	)
	if !ok {
		return nil
	}
	return cast
}

// Auth associated with this request
func (c Context) Auth() *Auth {
	var (
		raw      = c.Value(ctx.Auth)
		cast, ok = raw.(*Auth)
	)
	if !ok {
		return nil
	}
	return cast
}
