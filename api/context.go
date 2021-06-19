package api

import (
	"context"
	"net/http"
)

// custom context key type
type ctxKey int

const (
	ctxRequest ctxKey = iota
	ctxAuth
)

// Context is a simple wrapper around context.Context that adds helper methods
// to access additional information
type Context struct{ context.Context }

// Request current request
func (c Context) Request() *http.Request {
	var (
		raw      = c.Value(ctxRequest)
		cast, ok = raw.(*http.Request)
	)
	if !ok {
		return nil
	}
	return cast
}

// Auth associated with this request
func (c Context) Auth() *Auth {
	var (
		raw      = c.Value(ctxAuth)
		cast, ok = raw.(*Auth)
	)
	if !ok {
		return nil
	}
	return cast
}
