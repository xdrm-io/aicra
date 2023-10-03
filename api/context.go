package api

import (
	"net/http"

	"github.com/xdrm-io/aicra/internal/ctx"
)

// Extract the current internal data from a context.Context. Note: it never
// returns nil but struct fields can be nil
func Extract(r *http.Request) *Auth {
	c := ctx.Extract(r)
	if c == nil {
		return nil
	}
	auth, ok := c.(*Auth)
	if !ok {
		return nil
	}
	return auth
}
