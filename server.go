package aicra

import (
	"fmt"
	e "git.xdrm.io/go/aicra/err"
	"git.xdrm.io/go/aicra/internal/apirequest"
	"git.xdrm.io/go/aicra/internal/checker"
	"git.xdrm.io/go/aicra/internal/controller"
	"git.xdrm.io/go/aicra/middleware"
	"git.xdrm.io/go/aicra/response"
	"log"
	"net/http"
)

// Server represents an AICRA instance featuring:
// * its type checkers
// * its middlewares
// * its controllers (config)
type Server struct {
	controller *controller.Controller // controllers
	checker    *checker.Registry      // type checker registry
	middleware *middleware.Registry   // middlewares
}

// New creates a framework instance from a configuration file
func New(path string) (*Server, error) {

	/* (1) Init instance */
	var err error
	var i *Server

	/* (2) Load configuration */
	i.controller, err = controller.Load(path)
	if err != nil {
		return nil, err
	}

	/* (3) Default type registry */
	i.checker = checker.CreateRegistry(".build/type")

	/* (4) Default middleware registry */
	i.middleware = middleware.CreateRegistry(".build/middleware")

	return i, nil

}

// Listen binds the server to the given port
func (s *Server) Listen(port uint16) error {

	/* (1) Bind router */
	http.HandleFunc("/", s.manageRequest)

	/* (2) Bind listener */
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

}

// Router called for each request
func (s *Server) manageRequest(res http.ResponseWriter, httpReq *http.Request) {

	/* (1) Build request */
	req, err := apirequest.BuildFromHTTPRequest(httpReq)
	if err != nil {
		log.Fatal(err)
	}

	/* (2) Middleware: authentication */
	scope := s.middleware.Run(*httpReq)

	/* (3) Find a matching controller */
	controller := s.findController(req)
	if controller == nil {
		return
	}

	/* (4) Check if matching method exists */
	var method = controller.Method(httpReq.Method)

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
	parameters, paramError := s.extractParameters(req, method.Parameters)

	// Fail if argument check failed
	if paramError.Code != e.Success.Code {
		httpError(res, paramError)
		return
	}

	/* (5) Load controller
	---------------------------------------------------------*/
	callable, callErr := req.LoadController(httpReq.Method)
	if callErr.Code != e.Success.Code {
		httpError(res, callErr)
		log.Printf("[err] %s\n", err)
		return
	}

	/* (6) Execute and get response
	---------------------------------------------------------*/
	/* (1) Give Authorization header into controller */
	authHeader := httpReq.Header.Get("Authorization")
	if len(authHeader) > 0 {
		parameters["_AUTHORIZATION_"] = authHeader
	}

	/* (2) Give Scope into controller */
	parameters["_SCOPE_"] = scope

	/* (3) Execute */
	response := callable(parameters, response.New())

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
func (s *Server) extractParameters(req *apirequest.Request, methodParam map[string]*controller.Parameter) (map[string]interface{}, e.Error) {

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
