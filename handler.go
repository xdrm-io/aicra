package aicra

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/aicra/internal/ctx"
	"github.com/xdrm-io/aicra/internal/reqdata"
)

// Handler wraps the builder to handle requests
type Handler Builder

// ServeHTTP implements http.Handler and wraps it in middlewares (adapters)
func (s Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var h http.Handler = http.HandlerFunc(s.resolve)

	for _, mw := range s.middlewares {
		h = mw(h)
	}
	h.ServeHTTP(w, r)
}

// ServeHTTP implements http.Handler and wraps it in middlewares (adapters)
func (s Handler) resolve(w http.ResponseWriter, r *http.Request) {
	// match service from config
	var service = s.conf.Find(r)
	if service == nil {
		s.respond(w, nil, api.ErrUnknownService)
		return
	}

	// extract request data
	var input, err = extractInput(service, *r)
	if err != nil {
		if errors.Is(err, reqdata.ErrInvalidType) {
			s.respond(w, nil, api.ErrInvalidParam)
			return
		}
		s.respond(w, nil, api.ErrMissingParam)
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
		s.respond(w, nil, api.ErrUncallableService)
		return
	}

	// add info into context
	c := r.Context()
	c = context.WithValue(c, ctx.Request, r)
	c = context.WithValue(c, ctx.Response, w)
	c = context.WithValue(c, ctx.Auth, buildAuth(service.Scope, input.Data))

	// create http handler
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// should not happen
		auth := api.GetAuth(r.Context())
		if auth == nil {
			s.respond(w, nil, api.ErrForbidden)
			return
		}

		// reject non granted requests
		if !auth.Granted() {
			s.respond(w, nil, api.ErrForbidden)
			return
		}

		// execute the service handler
		s.handle(r.Context(), input, handler, service, w, r)
	})

	// run contextual middlewares
	for _, mw := range s.ctxMiddlewares {
		h = mw(h)
	}

	// serve using the pre-filled context
	h.ServeHTTP(w, r.WithContext(c))
}

// handle the service request with the associated handler func and respond using
// the handler func output
func (s *Handler) handle(c context.Context, input *reqdata.T, handler *serviceHandler, service *config.Service, w http.ResponseWriter, r *http.Request) {
	// pass execution to the handler function
	data, err := handler.dyn.Handle(c, input.Data)

	// rename data
	renamed := map[string]interface{}{}
	for key, value := range data {
		// find original name from 'rename' field
		for name, param := range service.Output {
			if param.Rename == key {
				renamed[name] = value
			}
		}
	}

	// write the http response
	s.respond(w, renamed, err)
}

func extractInput(service *config.Service, req http.Request) (*reqdata.T, error) {
	var dataset = reqdata.New(service)

	// URI data
	var err = dataset.GetURI(req)
	if err != nil {
		return nil, err
	}

	// query data
	err = dataset.GetQuery(req)
	if err != nil {
		return nil, err
	}

	// form/json data
	err = dataset.GetForm(req)
	if err != nil {
		return nil, err
	}

	return dataset, nil
}

// buildAuth builds the api.Auth struct from the service scope configuration
//
// it replaces format '[a]' in scope where 'a' is an existing input argument's
// name with its value
func buildAuth(scope [][]string, in map[string]interface{}) *api.Auth {
	updated := make([][]string, len(scope))

	// replace '[arg_name]' with the 'arg_name' value if it is a known variable
	// name
	for a, list := range scope {
		updated[a] = make([]string, len(list))
		for b, perm := range list {
			updated[a][b] = perm
			for name, value := range in {
				var (
					token       = fmt.Sprintf("[%s]", name)
					replacement = ""
				)
				if value != nil {
					replacement = fmt.Sprintf("[%v]", value)
				}
				updated[a][b] = strings.ReplaceAll(updated[a][b], token, replacement)
			}
		}
	}

	return &api.Auth{
		Required: updated,
		Active:   []string{},
	}
}
