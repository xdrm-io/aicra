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
func (server httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	/* (1) create api.Request from http.Request
	---------------------------------------------------------*/
	request, err := api.NewRequest(r)
	if err != nil {
		log.Fatal(err)
	}

	// 2. find a matching service for this path in the config
	serviceConf, pathIndex := server.config.Browse(request.URI)
	if serviceConf == nil {
		return
	}

	// 3. extract the service path from request URI
	servicePath := strings.Join(request.URI[:pathIndex], "/")
	if !strings.HasPrefix(servicePath, "/") {
		servicePath = "/" + servicePath
	}

	// 4. find method configuration from http method */
	var methodConf = serviceConf.Method(r.Method)
	if methodConf == nil {
		res := api.NewResponse(api.ErrorUnknownMethod())
		res.ServeHTTP(w, r)
		logError(res)
		return
	}

	// 5. parse data from the request (uri, query, form, json)
	data := reqdata.New(request.URI[pathIndex:], r)

	/* (2) check parameters
	---------------------------------------------------------*/
	parameters, paramError := server.extractParameters(data, methodConf.Parameters)

	// Fail if argument check failed
	if paramError.Code != api.ErrorSuccess().Code {
		res := api.NewResponse(paramError)
		res.ServeHTTP(w, r)
		logError(res)
		return
	}

	request.Param = parameters

	/* (3) search for the handler
	---------------------------------------------------------*/
	var foundHandler *api.Handler
	var found bool

	for _, handler := range server.handlers {
		if handler.GetPath() == servicePath {
			found = true
			if handler.GetMethod() == r.Method {
				foundHandler = handler
			}
		}
	}

	// fail if found no handler
	if foundHandler == nil {
		if found {
			res := api.NewResponse()
			res.SetError(api.ErrorUncallableMethod(), servicePath, r.Method)
			res.ServeHTTP(w, r)
			logError(res)
			return
		}

		res := api.NewResponse()
		res.SetError(api.ErrorUncallableService(), servicePath)
		res.ServeHTTP(w, r)
		logError(res)
		return
	}

	/* (4) execute handler and return response
	---------------------------------------------------------*/
	// 1. feed request with configuration scope
	request.Scope = methodConf.Scope

	// 2. execute
	res := api.NewResponse()
	foundHandler.Handle(*request, res)

	// 3. apply headers
	for key, values := range res.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 4. write to response
	res.ServeHTTP(w, r)
	return

}
