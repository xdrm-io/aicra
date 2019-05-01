package aicra

import (
	"log"
	"net/http"
	"os"
	"strings"

	"git.xdrm.io/go/aicra/api"

	"git.xdrm.io/go/aicra/internal/config"
	"git.xdrm.io/go/aicra/internal/reqdata"
	checker "git.xdrm.io/go/aicra/typecheck"
)

// Server represents an AICRA instance featuring: type checkers, services
type Server struct {
	services *config.Service
	Checkers *checker.Set
	handlers []*api.Handler
}

// New creates a framework instance from a configuration file
func New(configPath string) (*Server, error) {

	var err error

	// 1. init instance
	var i = &Server{
		services: nil,
		Checkers: checker.New(),
		handlers: make([]*api.Handler, 0),
	}

	// 2. open config file
	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	// 3. load configuration
	i.services, err = config.Parse(configFile)
	if err != nil {
		return nil, err
	}

	return i, nil

}

// ServeHTTP implements http.Handler and has to be called on each request
func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()

	// 1. build API request from HTTP request
	apiRequest, err := api.NewRequest(req)
	if err != nil {
		log.Fatal(err)
	}

	// 2. find a matching service for this path in the config
	serviceDef, pathIndex := s.services.Browse(apiRequest.URI)
	if serviceDef == nil {
		return
	}
	servicePath := strings.Join(apiRequest.URI[:pathIndex], "/")

	// 3. check if matching methodDef exists in config */
	var methodDef = serviceDef.Method(req.Method)
	if methodDef == nil {
		httpError(res, api.ErrorUnknownMethod())
		return
	}

	// 4. parse every input data from the request
	store := reqdata.New(apiRequest.URI[pathIndex:], req)

	/* (4) Check parameters
	---------------------------------------------------------*/
	parameters, paramError := s.extractParameters(store, methodDef.Parameters)

	// Fail if argument check failed
	if paramError.Code != api.ErrorSuccess().Code {
		httpError(res, paramError)
		return
	}

	apiRequest.Param = parameters

	/* (5) Search a matching handler
	---------------------------------------------------------*/
	var serviceHandler *api.Handler
	var serviceFound bool

	for _, handler := range s.handlers {
		if handler.GetPath() == servicePath {
			serviceFound = true
			if handler.GetMethod() == req.Method {
				serviceHandler = handler
			}
		}
	}

	// fail if found no handler
	if serviceHandler == nil {
		if serviceFound {
			apiError := api.ErrorUncallableMethod()
			apiError.Put(servicePath)
			apiError.Put(req.Method)
			httpError(res, apiError)
			return
		}

		apiError := api.ErrorUncallableService()
		apiError.Put(servicePath)
		httpError(res, apiError)
		return
	}

	/* (6) Execute handler and return response
	---------------------------------------------------------*/
	// 1. feed request with configuration scope
	apiRequest.Scope = methodDef.Permission

	// 1. execute
	apiResponse := api.NewResponse()
	serviceHandler.Handle(*apiRequest, apiResponse)

	// 2. apply headers
	for key, values := range apiResponse.Headers {
		for _, value := range values {
			res.Header().Add(key, value)
		}
	}

	// 3. build JSON apiResponse
	httpPrint(res, apiResponse)
	return

}

// HandleFunc sets a new handler for an HTTP method to a path
func (s *Server) HandleFunc(httpMethod, path string, handlerFunc api.HandlerFunc) {
	handler := api.NewHandler(httpMethod, path, handlerFunc)
	s.handlers = append(s.handlers, handler)
}

// Handle sets a new handler
func (s *Server) Handle(handler *api.Handler) {
	s.handlers = append(s.handlers, handler)
}
