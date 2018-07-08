package middleware

import (
	"net/http"
)

// Scope represents a list of scope processed by middlewares
// and used by the router to block/allow some uris
// it is also passed to controllers
type Scope []string

// Inspector updates the @Scope passed to it according to
// the @http.Request
type Inspector func(http.Request, *Scope)

// MiddleWare contains all necessary methods
// for a Middleware provided by user/developer
type MiddleWare struct {
	Inspect func(http.Request, *Scope)
}

// Registry represents a registry containing all registered
// middlewares to be processed before routing any request
type Registry struct {
	Middlewares []MiddleWare
}
