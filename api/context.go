package api

import (
	"context"
	"net/http"

	"github.com/xdrm-io/aicra/internal/ctx"
)

// Context defines the value stored in the request's context
type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Auth           *Auth
}

// Extract the current internal data from a context.Context. Note: it never
// returns nil but struct fields can be nil
func Extract(c context.Context) *Context {
	if c == nil {
		return &Context{}
	}
	var (
		raw      = c.Value(ctx.Key)
		cast, ok = raw.(*Context)
	)
	if !ok {
		return &Context{}
	}
	return cast
}
