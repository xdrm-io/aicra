package aicra

import (
	"errors"
	"git.xdrm.io/go/aicra/driver"
	e "git.xdrm.io/go/aicra/err"
	"git.xdrm.io/go/aicra/internal/api"
	"git.xdrm.io/go/aicra/internal/checker"
	"git.xdrm.io/go/aicra/internal/config"
	apirequest "git.xdrm.io/go/aicra/internal/request"
	"git.xdrm.io/go/aicra/middleware"
	"git.xdrm.io/go/aicra/response"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// Server represents an AICRA instance featuring:
// * its type checkers
// * its middlewares
// * its controllers (api config)
type Server struct {
	controller *api.Controller     // controllers
	checker    checker.Registry    // type checker registry
	middleware middleware.Registry // middlewares
	schema     *config.Schema
}

// New creates a framework instance from a configuration file
// _path is the json configuration path
// _driver is used to load/run the controllers and middlewares (default: )
//
func New(_path string) (*Server, error) {

	/* 1. Load config */
	schema, err := config.Parse("./aicra.json")
	if err != nil {
		return nil, err
	}

	/* 2. Init instance */
	var i = &Server{
		controller: nil,
		schema:     schema,
	}

	/* 3. Load configuration */
	i.controller, err = api.Parse(_path)
	if err != nil {
		return nil, err
	}

	/* 4. Load type registry */
	i.checker = checker.CreateRegistry()

	// add default types if set
	if schema.Types.Default {

		// driver is Plugin for defaults (even if generic for the controllers etc)
		defaultTypesDriver := new(driver.Plugin)
		files, err := filepath.Glob(filepath.Join(schema.Root, ".build/DEFAULT_TYPES/*.so"))
		if err != nil {
			return nil, errors.New("cannot load default types")
		}
		for _, path := range files {

			name := strings.TrimSuffix(filepath.Base(path), ".so")

			mwFunc, err := defaultTypesDriver.LoadChecker(path)
			if err != nil {
				log.Printf("cannot load default type checker '%s' | %s", name, err)
			}
			i.checker.Add(name, mwFunc)

		}
	}

	// add custom types
	for name, path := range schema.Types.Map {

		fullpath := schema.Driver.Build(schema.Root, schema.Types.Folder, path)
		mwFunc, err := schema.Driver.LoadChecker(fullpath)
		if err != nil {
			log.Printf("cannot load type checker '%s' | %s", name, err)
		}
		i.checker.Add(path, mwFunc)

	}

	/* 5. Load middleware registry */
	i.middleware = middleware.CreateRegistry()
	for name, path := range schema.Middlewares.Map {

		fullpath := schema.Driver.Build(schema.Root, schema.Middlewares.Folder, path)
		mwFunc, err := schema.Driver.LoadMiddleware(fullpath)
		if err != nil {
			log.Printf("cannot load middleware '%s' | %s", name, err)
		}
		i.middleware.Add(path, mwFunc)

	}

	return i, nil

}

// ServeHTTP implements http.Handler and has to be called on each request
func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()

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
	// get paths
	ctlBuildPath := strings.Join(apiRequest.Path, "/")
	ctlBuildPath = s.schema.Driver.Build(s.schema.Root, s.schema.Controllers.Folder, ctlBuildPath)

	// get controller
	ctlObject, err := s.schema.Driver.LoadController(ctlBuildPath)
	httpMethod := strings.ToUpper(req.Method)
	if err != nil {
		httpErr := e.UncallableController
		httpErr.Put(err)
		httpError(res, httpErr)
		log.Printf("err( %s )\n", err)
		return
	}

	var ctlMethod func(response.Arguments) response.Response
	// select method
	switch httpMethod {
	case "GET":
		ctlMethod = ctlObject.Get
	case "POST":
		ctlMethod = ctlObject.Post
	case "PUT":
		ctlMethod = ctlObject.Put
	case "DELETE":
		ctlMethod = ctlObject.Delete
	default:
		httpError(res, e.UnknownMethod)
		return
	}

	/* (6) Execute and get response
	---------------------------------------------------------*/
	/* (1) Give HTTP METHOD */
	parameters["_HTTP_METHOD_"] = httpMethod

	/* (2) Give Authorization header into controller */
	parameters["_AUTHORIZATION_"] = req.Header.Get("Authorization")

	/* (3) Give Scope into controller */
	parameters["_SCOPE_"] = scope

	/* (4) Execute */
	response := ctlMethod(parameters)

	/* (5) Extract http headers */
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
func (s *Server) extractParameters(req *apirequest.Request, methodParam map[string]*api.Parameter) (map[string]interface{}, e.Error) {

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
			err.Put(name)
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
			err.Put(param.Rename)
			err.Put("FILE")
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
			err.Put(param.Rename)
			err.Put(param.Type)
			err.Put(p.Value)
			break

		}

		parameters[param.Rename] = p.Value

	}

	return parameters, err
}
