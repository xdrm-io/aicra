package aicra

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/internal/config"
	"git.xdrm.io/go/aicra/internal/ctx"
	"git.xdrm.io/go/aicra/internal/reqdata"
)

// Handler wraps the builder to handle requests
type Handler Builder

// ServeHTTP implements http.Handler and wraps it in middlewares (adapters)
func (s Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var h = http.HandlerFunc(s.resolve)

	for _, adapter := range s.adapters {
		h = adapter(h)
	}
	h(w, r)
}

func (s Handler) resolve(w http.ResponseWriter, r *http.Request) {
	// 1. find a matching service from config
	var service = s.conf.Find(r)
	if service == nil {
		handleError(api.ErrUnknownService, w, r)
		return
	}

	// 2. extract request data
	var input, err = extractInput(service, *r)
	if err != nil {
		handleError(api.ErrMissingParam, w, r)
		return
	}

	// 3. find a matching handler
	var handler *apiHandler
	for _, h := range s.handlers {
		if h.Method == service.Method && h.Path == service.Pattern {
			handler = h
		}
	}

	// 4. fail on no matching handler
	if handler == nil {
		handleError(api.ErrUncallableService, w, r)
		return
	}

	// replace format '[a]' in scope where 'a' is an existing input's name
	scope := make([][]string, len(service.Scope))
	for a, list := range service.Scope {
		scope[a] = make([]string, len(list))
		for b, perm := range list {
			scope[a][b] = perm
			for name, value := range input.Data {
				var (
					token       = fmt.Sprintf("[%s]", name)
					replacement = ""
				)
				if value != nil {
					replacement = fmt.Sprintf("[%v]", value)
				}
				scope[a][b] = strings.ReplaceAll(scope[a][b], token, replacement)
			}
		}
	}

	var auth = api.Auth{
		Required: scope,
		Active:   []string{},
	}

	// 5. run auth-aware middlewares
	var h = api.AuthHandlerFunc(func(a api.Auth, w http.ResponseWriter, r *http.Request) {
		if !a.Granted() {
			handleError(api.ErrPermission, w, r)
			return
		}

		s.handle(input, handler, service, w, r)
	})

	for _, adapter := range s.authAdapters {
		h = adapter(h)
	}
	h(auth, w, r)

}

func (s *Handler) handle(input *reqdata.T, handler *apiHandler, service *config.Service, w http.ResponseWriter, r *http.Request) {
	// build context with builtin data
	c := r.Context()
	c = context.WithValue(c, ctx.Request, r)
	c = context.WithValue(c, ctx.Response, w)
	c = context.WithValue(c, ctx.Auth, w)
	apictx := &api.Context{Context: c}

	// pass execution to the handler
	var outData, outErr = handler.dyn.Handle(apictx, input.Data)

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
