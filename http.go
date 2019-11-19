package aicra

import (
	"log"
	"net/http"
	"strings"

	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/internal/reqdata"
)

// httpServer wraps the aicra server to allow handling http requests
type httpServer Server

// ServeHTTP implements http.Handler and has to be called on each request
func (s httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// 1. build API request from HTTP request
	request, err := api.NewRequest(r)
	if err != nil {
		log.Fatal(err)
	}

	// 2. find a matching service for this path in the config
	serviceDef, pathIndex := s.services.Browse(request.URI)
	if serviceDef == nil {
		return
	}
	servicePath := strings.Join(request.URI[:pathIndex], "/")
	if !strings.HasPrefix(servicePath, "/") {
		servicePath = "/" + servicePath
	}

	// 3. check if matching methodDef exists in config */
	var methodDef = serviceDef.Method(r.Method)
	if methodDef == nil {
		response := api.NewResponse(api.ErrorUnknownMethod())
		response.ServeHTTP(w, r)
		logError(response)
		return
	}

	// 4. parse every input data from the request
	store := reqdata.New(request.URI[pathIndex:], r)

	/* (4) Check parameters
	---------------------------------------------------------*/
	parameters, paramError := s.extractParameters(store, methodDef.Parameters)

	// Fail if argument check failed
	if paramError.Code != api.ErrorSuccess().Code {
		response := api.NewResponse(paramError)
		response.ServeHTTP(w, r)
		logError(response)
		return
	}

	request.Param = parameters

	/* (5) Search a matching handler
	---------------------------------------------------------*/
	var serviceHandler *api.Handler
	var serviceFound bool

	for _, handler := range s.handlers {
		if handler.GetPath() == servicePath {
			serviceFound = true
			if handler.GetMethod() == r.Method {
				serviceHandler = handler
			}
		}
	}

	// fail if found no handler
	if serviceHandler == nil {
		if serviceFound {
			response := api.NewResponse()
			response.SetError(api.ErrorUncallableMethod(), servicePath, r.Method)
			response.ServeHTTP(w, r)
			logError(response)
			return
		}

		response := api.NewResponse()
		response.SetError(api.ErrorUncallableService(), servicePath)
		response.ServeHTTP(w, r)
		logError(response)
		return
	}

	/* (6) Execute handler and return response
	---------------------------------------------------------*/
	// 1. feed request with configuration scope
	request.Scope = methodDef.Scope

	// 1. execute
	response := api.NewResponse()
	serviceHandler.Handle(*request, response)

	// 2. apply headers
	for key, values := range response.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 3. write to response
	response.ServeHTTP(w, r)
	return

}
