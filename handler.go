package aicra

import (
	"context"
	"net/http"

	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/internal/ctx"
	"github.com/xdrm-io/aicra/runtime"
)

// Handler wraps the builder to handle requests
type Handler Builder

// ServeHTTP implements http.Handler and wraps it in middlewares (adapters)
func (s Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.uriLimit > 0 && len(r.URL.RequestURI()) > s.uriLimit {
		runtime.Respond(w, nil, api.ErrURITooLong)
		return
	}
	if s.bodyLimit > 0 && r.ContentLength > s.bodyLimit {
		runtime.Respond(w, nil, api.ErrBodyTooLarge)
		return
	}

	var h http.Handler = http.HandlerFunc(s.resolve)

	for _, mw := range s.middlewares {
		h = mw(h)
	}
	h.ServeHTTP(w, r)
}

// ServeHTTP implements http.Handler and wraps it in middlewares
func (s Handler) resolve(w http.ResponseWriter, r *http.Request) {
	// match service from config
	var service = s.conf.Find(r, s.validators)
	if service == nil {
		runtime.Respond(w, nil, api.ErrUnknownService)
		return
	}

	// match handler
	var handler *serviceHandler
	for _, h := range s.handlers {
		if h.Method == service.Method && h.Path == service.Pattern {
			handler = h
		}
	}

	// no handler found
	if handler == nil {
		// should never fail as the builder ensures all services are plugged
		// properly
		runtime.Respond(w, nil, api.ErrUncallableService)
		return
	}

	// add info into context
	c := context.WithValue(r.Context(), ctx.Key, &api.Context{
		Auth: &api.Auth{
			Required: service.Scope,
			Active:   make([]string, 0),
		},
	})

	// run contextual middlewares
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := api.Extract(r.Context())

		// reject non granted requests
		if !ctx.Auth.Granted() {
			runtime.Respond(w, nil, api.ErrForbidden)
			return
		}

		// execute the service handler
		handler.fn.ServeHTTP(w, r)
	})
	for _, mw := range s.ctxMiddlewares {
		h = mw(h)
	}

	// serve using the pre-filled context
	h.ServeHTTP(w, r.WithContext(c))
}
