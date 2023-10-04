package ctx

import (
	"net/http"
	"sync"
)

// contexts contains every internal context for every http request (by addr)
var (
	contexts = make(map[*http.Request]any, 0)
	mu       sync.RWMutex
)

// Context defines the value stored in the request's context
type Context any

// Register a new context for a given http request
func Register(addr *http.Request, v any) {
	if addr == nil {
		return
	}
	mu.Lock()
	contexts[addr] = v
	mu.Unlock()
}

// Release a context for a given http request
func Release(addr *http.Request) {
	mu.Lock()
	delete(contexts, addr)
	mu.Unlock()
}

// Extract the current internal data from a context.Context. Note: it never
// returns nil but struct fields can be nil
func Extract(addr *http.Request) Context {
	mu.RLock()
	ctx, _ := contexts[addr]
	mu.RUnlock()
	return ctx
}
