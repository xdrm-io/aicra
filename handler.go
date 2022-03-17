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
	if s.uriLimit > 0 && len(r.URL.RequestURI()) > s.uriLimit {
		s.respond(w, nil, api.ErrURITooLong)
		return
	}
	if s.bodyLimit > 0 && r.ContentLength > s.bodyLimit {
		s.respond(w, nil, api.ErrBodyTooLarge)
		return
	}

	var h http.Handler = http.HandlerFunc(s.resolve)

	for _, mw := range s.middlewares {
		h = mw(h)
	}
	h.ServeHTTP(w, r)
}

var zeroRequest http.Request

// ServeHTTP implements http.Handler and wraps it in middlewares (adapters)
func (s Handler) resolve(w http.ResponseWriter, r *http.Request) {
	// match service from config
	var service = s.conf.Find(r)
	if service == nil {
		s.respond(w, nil, api.ErrUnknownService)
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
		s.respond(w, nil, api.ErrUncallableService)
		return
	}

	// start building the input but only URI parameters for now.
	// They might be required to build parametric authorization c.f. buildAuth()
	// Only URI arguments can be used
	var input = reqdata.NewRequest(service)
	if err := input.ExtractURI(r); err != nil {
		// should never fail as type validators are always checked in
		// s.conf.Find -> config.Service.matchPattern
		input.Release()
		s.respond(w, nil, enrichInputError(err))
		return
	}

	// add info into context
	c := context.WithValue(r.Context(), ctx.Key, &api.Context{
		Request:        r,
		ResponseWriter: w,
		Auth:           buildAuth(service.Scope, service.ScopeVars, input.Data),
	})

	// create http handler
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := api.Extract(r.Context())
		if ctx == nil || ctx.Auth == nil {
			// should never happen
			input.Release()
			s.respond(w, nil, api.ErrForbidden)
			return
		}

		// reject non granted requests
		if !ctx.Auth.Granted() {
			input.Release()
			s.respond(w, nil, api.ErrForbidden)
			return
		}

		// extract remaining input parameters
		if err := input.ExtractQuery(ctx.Request); err != nil {
			input.Release()
			s.respond(w, nil, enrichInputError(err))
			return
		}
		if err := input.ExtractForm(ctx.Request); err != nil {
			input.Release()
			s.respond(w, nil, enrichInputError(err))
			return
		}

		// execute the service handler
		s.handle(r.Context(), input, handler, service, w)
		input.Release()
	})

	// run contextual middlewares
	for _, mw := range s.ctxMiddlewares {
		h = mw(h)
	}

	// serve using the pre-filled context
	h.ServeHTTP(w, zeroRequest.WithContext(c))
}

// handle the service request with the associated handler func and respond using
// the handler func output
func (s *Handler) handle(c context.Context, input *reqdata.Request, handler *serviceHandler, service *config.Service, w http.ResponseWriter) {
	// pass execution to the handler function
	data, err := handler.dyn.Handle(c, input.Data)

	// rename data
	renamed := make(map[string]interface{}, len(service.Output))
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

// enrichInputError parses and manages the input error to add field information
func enrichInputError(err error) error {
	if err == nil {
		return nil
	}

	// invalid data according to its validator
	if errors.Is(err, reqdata.ErrInvalidType) {
		cast, ok := err.(*reqdata.Err)
		if !ok {
			return api.ErrInvalidParam
		}

		// add field name to error
		return api.Error(
			api.ErrInvalidParam.Status(),
			fmt.Errorf("%s: %w", cast.Field(), api.ErrInvalidParam),
		)
	}

	var (
		missingParam    = errors.Is(err, reqdata.ErrMissingRequiredParam)
		missingURIParam = errors.Is(err, reqdata.ErrMissingURIParameter)
	)
	if missingParam || missingURIParam {
		cast, ok := err.(*reqdata.Err)
		if !ok {
			return api.ErrMissingParam
		}
		// add field name to error
		return api.Error(
			api.ErrMissingParam.Status(),
			fmt.Errorf("%s: %w", cast.Field(), api.ErrMissingParam),
		)
	}

	return api.ErrMissingParam
}

// buildAuth builds the api.Auth struct from the service scope configuration
//
// it replaces format '[a]' in scope where 'a' is an existing input argument's
// name with its value.
// Warning notice: only uri parameters are allowed
func buildAuth(scope [][]string, scopeVars []config.ScopeVar, in map[string]interface{}) *api.Auth {
	if len(scope) < 1 || len(scopeVars) < 1 {
		return &api.Auth{Required: scope}
	}

	// copy scope
	updated := make([][]string, len(scope))
	for a, list := range scope {
		updated[a] = make([]string, len(list))
		for b, perm := range list {
			updated[a][b] = perm
		}
	}

	// replace '[arg_name]' with the 'arg_name' value if it is a known variable
	// name
	for _, sv := range scopeVars {
		value, set := in[sv.CaptureName]
		if !set {
			continue
		}
		a, b := sv.Position[0], sv.Position[1]
		updated[a][b] = strings.ReplaceAll(
			updated[a][b],
			fmt.Sprintf("[%s]", sv.CaptureName),
			fmt.Sprintf("[%v]", value),
		)
	}
	return &api.Auth{Required: updated}
}
