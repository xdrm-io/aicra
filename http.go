package aicra

import (
	"log"
	"net/http"

	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/internal/reqdata"
)

// httpHandler wraps the aicra server to allow handling http requests
type httpHandler Server

// ServeHTTP implements http.Handler and has to be called on each request
func (server httpHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	// 1. find a matching service in the config
	service := server.config.Find(req)
	if service == nil {
		response := api.EmptyResponse().WithError(api.ErrorUnknownService)
		response.ServeHTTP(res, req)
		logError(response)
		return
	}

	// 2. build input parameter receiver
	dataset := reqdata.New(service)

	// 3. extract URI data
	err := dataset.ExtractURI(req)
	if err != nil {
		response := api.EmptyResponse().WithError(api.ErrorMissingParam)
		response.ServeHTTP(res, req)
		logError(response)
		return
	}

	// 4. extract query data
	err = dataset.ExtractQuery(req)
	if err != nil {
		response := api.EmptyResponse().WithError(api.ErrorMissingParam)
		response.ServeHTTP(res, req)
		logError(response)
		return
	}

	// 5. extract form/json data
	err = dataset.ExtractForm(req)
	if err != nil {
		response := api.EmptyResponse().WithError(api.ErrorMissingParam)
		response.ServeHTTP(res, req)
		logError(response)
		return
	}

	// 6. find a matching handler
	var foundHandler *handler
	var found bool

	for _, handler := range server.handlers {
		if handler.Method == service.Method && handler.Path == service.Pattern {
			foundHandler = handler
			found = true
		}
	}

	// 7. fail if found no handler
	if foundHandler == nil {
		if found {
			r := api.EmptyResponse().WithError(api.ErrorUncallableService)
			r.ServeHTTP(res, req)
			logError(r)
			return
		}

		r := api.EmptyResponse().WithError(api.ErrorUnknownService)
		r.ServeHTTP(res, req)
		logError(r)
		return
	}

	// 8. build api.Request from http.Request
	apireq, err := api.NewRequest(req)
	if err != nil {
		log.Fatal(err)
	}

	// 9. feed request with scope & parameters
	apireq.Scope = service.Scope
	apireq.Param = dataset.Data

	// 10. execute
	returned, apiErr := foundHandler.dynHandler.Handle(dataset.Data)
	response := api.EmptyResponse().WithError(apiErr)
	for key, value := range returned {

		// find original name from rename
		for name, param := range service.Output {
			if param.Rename == key {
				response.SetData(name, value)
			}
		}
	}

	// 11. apply headers
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	for key, values := range response.Headers {
		for _, value := range values {
			res.Header().Add(key, value)
		}
	}

	// 12. write to response
	response.ServeHTTP(res, req)
}
