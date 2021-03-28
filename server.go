package aicra

import (
	"net/http"

	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/internal/config"
	"git.xdrm.io/go/aicra/internal/reqdata"
)

// Server hides the builder and allows handling http requests
type Server Builder

// ServeHTTP implements http.Handler and is called on each request
func (server Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	// 1. find a matching service in the config
	service := server.conf.Find(req)
	if service == nil {
		handleError(api.ErrUnknownService, w, r)
		return
	}

	// 2. extract request data
	dataset, err := extractRequestData(service, *req)
	if err != nil {
		handleError(api.ErrMissingParam, w, r)
		return
	}

	// 3. find a matching handler
	var handler *apiHandler
	for _, h := range server.handlers {
		if h.Method == service.Method && h.Path == service.Pattern {
			handler = h
		}
	}

	// 4. fail if found no handler
	if handler == nil {
		handleError(api.ErrUncallableService, w, r)
		return
	}

	// 5. execute
	returned, apiErr := handler.dyn.Handle(dataset.Data)

	// 6. build response from returned data
	response := api.EmptyResponse().WithError(apiErr)
	for key, value := range returned {

		// find original name from rename
		for name, param := range service.Output {
			if param.Rename == key {
				response.SetData(name, value)
			}
		}
	}

	// 7. apply headers
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	for key, values := range response.Headers {
		for _, value := range values {
			res.Header().Add(key, value)
		}
	}

	response.ServeHTTP(res, req)
}

func handleError(err api.Err, w http.ResponseWriter, r *http.Request) {
	var response = api.EmptyResponse().WithError(err)
	response.ServeHTTP(w, r)
}

func extractRequestData(service *config.Service, req http.Request) (*reqdata.T, error) {
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
