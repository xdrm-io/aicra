package aicra

import (
	"log"
	"net/http"

	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/internal/reqdata"
)

// httpServer wraps the aicra server to allow handling http requests
type httpServer Server

// ServeHTTP implements http.Handler and has to be called on each request
func (server httpServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	// 1. find a matching service in the config
	service := server.config.Find(req)
	if service == nil {
		response := api.NewResponse(api.ErrorUnknownService())
		response.ServeHTTP(res, req)
		logError(response)
		return
	}

	// 2. build input parameter receiver
	dataset := reqdata.New(service)

	// 3. extract URI data
	err := dataset.ExtractURI(req)
	if err != nil {
		response := api.NewResponse(api.ErrorMissingParam())
		response.ServeHTTP(res, req)
		logError(response)
		return
	}

	// 4. extract query data
	err = dataset.ExtractQuery(req)
	if err != nil {
		response := api.NewResponse(api.ErrorMissingParam())
		response.ServeHTTP(res, req)
		logError(response)
		return
	}

	// 5. extract form/json data
	err = dataset.ExtractForm(req)
	if err != nil {
		response := api.NewResponse(api.ErrorMissingParam())
		response.ServeHTTP(res, req)
		logError(response)
		return
	}

	// 6. find a matching handler
	var foundHandler *api.Handler
	var found bool

	for _, handler := range server.handlers {
		if handler.GetMethod() == service.Method && handler.GetPath() == service.Pattern {
			found = true
		}
	}

	// 7. fail if found no handler
	if foundHandler == nil {
		if found {
			r := api.NewResponse()
			r.SetError(api.ErrorUncallableService(), service.Method, service.Pattern)
			r.ServeHTTP(res, req)
			logError(r)
			return
		}

		r := api.NewResponse()
		r.SetError(api.ErrorUnknownService(), service.Method, service.Pattern)
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
	response := api.NewResponse()
	foundHandler.Handle(*apireq, response)

	// 11. apply headers
	for key, values := range response.Headers {
		for _, value := range values {
			res.Header().Add(key, value)
		}
	}

	// 12. write to response
	response.ServeHTTP(res, req)
}
