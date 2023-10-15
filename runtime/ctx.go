package runtime

import (
	"net/http"

	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/internal/ctx"
)

// Context defines per-request values
type Context struct {
	// Auth contains the authentication information
	Auth *api.Auth
}

// GetAuth returns the authentication information from the request
func GetAuth(r *http.Request) *api.Auth {
	raw := ctx.Extract(r)
	if raw == nil {
		return nil
	}
	c, ok := raw.(*Context)
	if !ok || c == nil {
		return nil
	}
	return c.Auth
}
