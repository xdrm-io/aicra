package middleware

import (
	"net/http"
)

// Scope represents a list of scope processed by middlewares
// and used by the router to block/allow some uris
// it is also passed to controllers
//
// DISCLAIMER: it is used to help developers but for compatibility
//             purposes, the type is always used as its definition ([]string)
type Scope []string

type MiddlewareFunc func(http.Request, *[]string)

// Middleware updates the @Scope passed to it according to
// the @http.Request
type Middleware interface {
	Inspect(http.Request, *[]string)
}

// Wrapper is a struct that stores middleware Inspect() method
type Wrapper struct {
	Inspect func(http.Request, *[]string)
}

// Registry represents a registry containing all registered
// middlewares to be processed before routing any request
type Registry struct {
	Middlewares []*Wrapper
}
