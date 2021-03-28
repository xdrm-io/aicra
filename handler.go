package aicra

import (
	"net/http"

	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/internal/config"
	"git.xdrm.io/go/aicra/internal/reqdata"
)

// Handler wraps the builder to handle requests
type Handler Builder

// ServeHTTP implements http.Handler
func (s Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

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

	// 5. pass execution to the handler
	var outData, outErr = handler.dyn.Handle(input.Data)

	// 6. build res from returned data
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
