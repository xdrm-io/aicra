package gfw

import (
	"encoding/json"
	"fmt"
	"git.xdrm.io/go/aicra/checker"
	"git.xdrm.io/go/aicra/config"
	"git.xdrm.io/go/aicra/err"
	"git.xdrm.io/go/aicra/implement"
	"git.xdrm.io/go/aicra/request"
	"log"
	"net/http"
)

// Init initilises a new framework instance
//
// - path is the configuration path
//
// - if typeChecker is nil, defaults will be used (all *.so files
//   inside ./types local directory)
func Init(path string, typeChecker *checker.TypeRegistry) (*Server, error) {

	/* (1) Init instance */
	inst := &Server{
		config: nil,
		Params: make(map[string]interface{}),
	}

	/* (2) Load configuration */
	config, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	inst.config = config

	/* (3) Store registry if not nil */
	if typeChecker != nil {
		inst.Checker = typeChecker
		return inst, nil
	}

	/* (4) Default registry creation */
	inst.Checker = checker.CreateRegistry(true)

	return inst, nil
}

// Listens and binds the server to the given port
func (s *Server) Launch(port uint16) error {

	/* (1) Bind router */
	http.HandleFunc("/", s.routeRequest)

	/* (2) Bind listener */
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

}

// Router called for each request
func (s *Server) routeRequest(res http.ResponseWriter, httpReq *http.Request) {

	/* (1) Build request */
	req, err2 := request.Build(httpReq)
	if err2 != nil {
		log.Fatal(err2)
	}

	/* (2) Middleware: authentication */
	// TODO: Auth

	/* (3) Find a matching controller */
	controller := s.findController(req)

	/* (4) Check if matching method exists */
	var method = controller.Method(httpReq.Method)

	if method == nil {
		httpError(res, err.UnknownMethod)
		return
	}

	/* (4) Check parameters
	---------------------------------------------------------*/
	var paramError err.Error = err.Success
	parameters := make(map[string]interface{})
	for name, param := range method.Parameters {

		/* (1) Extract value */
		p, isset := req.Data.Set[name]

		/* (2) Required & missing */
		if !isset && !param.Optional {
			paramError = err.MissingParam
			paramError.BindArgument(name)
			break
		}

		/* (3) Optional & missing: set default value */
		if !isset {
			p = &request.Parameter{
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
			paramError = err.InvalidParam
			paramError.BindArgument(param.Rename)
			paramError.BindArgument("FILE")
			break
		}

		/* (6) Do not check if file */
		if gotFile {
			parameters[param.Rename] = p.Value
			continue
		}

		/* (7) Check type */
		if s.Checker.Run(param.Type, p.Value) != nil {

			paramError = err.InvalidParam
			paramError.BindArgument(param.Rename)
			paramError.BindArgument(param.Type)
			paramError.BindArgument(p.Value)
			break

		}

		parameters[param.Rename] = p.Value

	}

	// Fail if argument check failed
	if paramError.Code != err.Success.Code {
		httpError(res, paramError)
		return
	}

	/* (5) Load controller
	---------------------------------------------------------*/
	callable, err := req.LoadController(httpReq.Method)
	if err != nil {
		log.Printf("[err] %s\n", err)
		return
	}

	/* (6) Execute and get response
	---------------------------------------------------------*/
	/* (1) Show Authorization header into controller */
	authHeader := httpReq.Header.Get("Authorization")
	if len(authHeader) > 0 {
		parameters["_AUTHORIZATION_"] = authHeader
	}

	/* (2) Execute */
	responseBarebone := implement.NewResponse()
	response := callable(parameters, responseBarebone)

	/* (3) Extract http headers */
	for k, v := range response.Dump() {
		if k == "_REDIRECT_" {
			if newLocation, ok := v.(string); ok {
				httpRedirect(res, newLocation)
			}
			continue
		}
	}

	/* (4) Build JSON response */
	httpPrint(res, response)
	return

}
