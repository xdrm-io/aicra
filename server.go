package aicra

import (
	"errors"
	"git.xdrm.io/go/aicra/driver"
	e "git.xdrm.io/go/aicra/err"
	"git.xdrm.io/go/aicra/internal/checker"
	"git.xdrm.io/go/aicra/internal/config"
	apirequest "git.xdrm.io/go/aicra/internal/request"
	"git.xdrm.io/go/aicra/middleware"
	"log"
	"net/http"
)

// Server represents an AICRA instance featuring:
// * its type checkers
// * its middlewares
// * its controllers (config)
type Server struct {
	controller *config.Controller   // controllers
	checker    *checker.Registry    // type checker registry
	middleware *middleware.Registry // middlewares
	driver     driver.Driver
}

// ErrNilDriver is raised when a NULL driver is given to the constructor
var ErrNilDriver = errors.New("the driver is <nil>")

// New creates a framework instance from a configuration file
func New(_path string, _driver driver.Driver) (*Server, error) {

	if _driver == nil {
		return nil, ErrNilDriver
	}

	/* (1) Init instance */
	var err error
	var i = &Server{
		controller: nil,
		driver:     _driver,
	}

	/* (2) Load configuration */
	i.controller, err = config.Load(_path)
	if err != nil {
		return nil, err
	}

	/* (3) Default type registry */
	i.checker = checker.CreateRegistry(".build/type")

	/* (4) Default middleware registry */
	i.middleware = middleware.CreateRegistry(".build/middleware")

	return i, nil

}

// ServeHTTP implements http.Handler and has to be called on each request
func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	/* (1) Build request */
	apiRequest, err := apirequest.FromHTTP(req)
	if err != nil {
		log.Fatal(err)
	}

	/* (2) Launch middlewares to build the scope */
	scope := s.middleware.Run(*req)

	/* (3) Find a matching controller */
	controller := s.matchController(apiRequest)
	if controller == nil {
		return
	}

	/* (4) Check if matching method exists */
	var method = controller.Method(req.Method)

	if method == nil {
		httpError(res, e.UnknownMethod)
		return
	}

	/* (5) Check scope permissions */
	if !method.CheckScope(scope) {
		httpError(res, e.Permission)
		return
	}

	/* (4) Check parameters
	---------------------------------------------------------*/
	parameters, paramError := s.extractParameters(apiRequest, method.Parameters)

	// Fail if argument check failed
	if paramError.Code != e.Success.Code {
		httpError(res, paramError)
		return
	}

	/* (5) Load controller
	---------------------------------------------------------*/
	controllerImplementation, callErr := apiRequest.LoadController(req.Method, s.driver)
	if callErr.Code != e.Success.Code {
		httpError(res, callErr)
		log.Printf("[err] %s\n", err)
		return
	}

	/* (6) Execute and get response
	---------------------------------------------------------*/
	/* (1) Give Authorization header into controller */
	parameters["_AUTHORIZATION_"] = req.Header.Get("Authorization")

	/* (2) Give Scope into controller */
	parameters["_SCOPE_"] = scope

	/* (3) Execute */
	response := controllerImplementation(parameters)

	/* (4) Extract http headers */
	for k, v := range response.Dump() {
		if k == "_REDIRECT_" {
			if newLocation, ok := v.(string); ok {
				httpRedirect(res, newLocation)
			}
			continue
		}
	}

	/* (5) Build JSON response */
	httpPrint(res, response)
	return

}

// extractParameters extracts parameters for the request and checks
// every single one according to configuration options
func (s *Server) extractParameters(req *apirequest.Request, methodParam map[string]*config.Parameter) (map[string]interface{}, e.Error) {

	// init vars
	err := e.Success
	parameters := make(map[string]interface{})

	// for each param of the config
	for name, param := range methodParam {

		/* (1) Extract value */
		p, isset := req.Data.Set[name]

		/* (2) Required & missing */
		if !isset && !param.Optional {
			err = e.MissingParam
			err.BindArgument(name)
			return nil, err
		}

		/* (3) Optional & missing: set default value */
		if !isset {
			p = &apirequest.Parameter{
				Parsed: true,
				File:   param.Type == "FILE",
				Value:  nil,
			}
			if param.Default != nil {
				p.Value = *param.Default
			}

			// we are done
			parameters[param.Rename] = p.Value
			continue
		}

		/* (4) Parse parameter if not file */
		if !p.File {
			p.Parse()
		}

		/* (5) Fail on unexpected multipart file */
		waitFile, gotFile := param.Type == "FILE", p.File
		if gotFile && !waitFile || !gotFile && waitFile {
			err = e.InvalidParam
			err.BindArgument(param.Rename)
			err.BindArgument("FILE")
			return nil, err
		}

		/* (6) Do not check if file */
		if gotFile {
			parameters[param.Rename] = p.Value
			continue
		}

		/* (7) Check type */
		if s.checker.Run(param.Type, p.Value) != nil {

			err = e.InvalidParam
			err.BindArgument(param.Rename)
			err.BindArgument(param.Type)
			err.BindArgument(p.Value)
			break

		}

		parameters[param.Rename] = p.Value

	}

	return parameters, err
}
