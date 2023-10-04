package aicra

import (
	"net/http"

	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/internal/config"
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
	fragments := config.URIFragments(r.URL.Path)

	// match endpoint from config
	endpoint := s.conf.Find(r.Method, fragments, s.validators)
	if endpoint == nil {
		runtime.Respond(w, nil, api.ErrUnknownService)
		return
	}

	// match handler
	handler, ok := s.handlers[endpoint.Name]
	if !ok || handler == nil {
		// should never fail as the builder ensures all services are plugged
		// properly
		runtime.Respond(w, nil, api.ErrUncallableService)
		return
	}

	// add info into context
	ctx.Register(r, &runtime.Context{
		Fragments: fragments,
		Auth: &api.Auth{
			Required: endpoint.Scope,
			Active:   nil,
		},
	})

	// run contextual middlewares
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := runtime.GetAuth(r)

		// reject non granted requests
		if !auth.Granted() {
			runtime.Respond(w, nil, api.ErrForbidden)
			return
		}

		// execute the service handler
		handler.ServeHTTP(w, r)
	})
	for _, mw := range s.ctxMiddlewares {
		h = mw(h)
	}

	// serve using the pre-filled context
	h.ServeHTTP(w, r)
	ctx.Release(r)
}
