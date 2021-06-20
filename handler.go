package aicra

import (
	"context"
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
	// 1ind a matching service from config
	var service = s.conf.Find(r)
	if service == nil {
		handleError(api.ErrUnknownService, w, r)
		return
	}

	// extract request data
	var input, err = extractInput(service, *r)
	if err != nil {
		handleError(api.ErrMissingParam, w, r)
		return
	}

	// find a matching handler
	var handler *apiHandler
	for _, h := range s.handlers {
		if h.Method == service.Method && h.Path == service.Pattern {
			handler = h
		}
	}

	// fail on no matching handler
	if handler == nil {
		handleError(api.ErrUncallableService, w, r)
		return
	}

	// build context with builtin data
	c := r.Context()
	c = context.WithValue(c, ctx.Request, r)
	c = context.WithValue(c, ctx.Response, w)
	c = context.WithValue(c, ctx.Auth, buildAuth(service.Scope, input.Data))

	// create http handler
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := api.GetAuth(r.Context())
		if auth == nil {
			handleError(api.ErrPermission, w, r)
			return
		}

		// reject non granted requests
		if !auth.Granted() {
			handleError(api.ErrPermission, w, r)
			return
		}

		// use context defined in the request
		s.handle(r.Context(), input, handler, service, w, r)
	})

	// run middlewares the handler
	for _, mw := range s.ctxMiddlewares {
		h = mw(h)
	}

	// serve using the context with values
	h.ServeHTTP(w, r.WithContext(c))
}

func (s *Handler) handle(c context.Context, input *reqdata.T, handler *apiHandler, service *config.Service, w http.ResponseWriter, r *http.Request) {
	// pass execution to the handler
	var outData, outErr = handler.dyn.Handle(c, input.Data)

	// build response from returned arguments
	var res = api.EmptyResponse().WithError(outErr)
	for key, value := range outData {

		// find original name from 'rename' field
		for name, param := range service.Output {
			if param.Rename == key {
				res.SetData(name, value)
			}
		}
	}

	// 7. apply headers
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	for key, values := range res.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	res.ServeHTTP(w, r)
}

func handleError(err api.Err, w http.ResponseWriter, r *http.Request) {
	var response = api.EmptyResponse().WithError(err)
	response.ServeHTTP(w, r)
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
