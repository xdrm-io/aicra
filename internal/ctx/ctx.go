package ctx

import (
	"net/http"
	"sync"
)

// contexts contains every internal context for every http request (by addr)
var contexts = sync.Map{
	// key: *http.Request
	// value: Context
}

// Context defines the value stored in the request's context
type Context any

// Register a new context for a given http request
func Register(addr *http.Request, v any) {
	if addr == nil {
		return
	}
	contexts.LoadOrStore(addr, v)
}

// Release a context for a given http request
func Release(addr *http.Request) {
	contexts.Delete(addr)
}

// Extract the current internal data from a context.Context. Note: it never
// returns nil but struct fields can be nil
func Extract(addr *http.Request) Context {
	ctx, _ := contexts.Load(addr)
	return ctx
}
