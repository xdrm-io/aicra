package gfw

import (
	"fmt"
	"git.xdrm.io/xdrm-brackets/gfw/config"
	"git.xdrm.io/xdrm-brackets/gfw/err"
	"log"
	"net/http"
	"strings"
)

func (s *Server) route(res http.ResponseWriter, req *http.Request) {

	/* (1) Build request
	---------------------------------------------------------*/
	/* (1) Try to build request */
	request, err2 := buildRequest(req)
	if err2 != nil {
		log.Fatal(req)
	}

	/* (2) Find a controller
	---------------------------------------------------------*/
	controller := s.findController(request)

	/* (3) Check method
	---------------------------------------------------------*/
	method := s.getMethod(controller, req.Method)

	if method == nil {
		Json, _ := err.UnknownMethod.MarshalJSON()
		res.Header().Add("Content-Type", "application/json")
		res.Write(Json)
		log.Printf("[err] %s\n", err.UnknownMethod.Reason)
		return
	}

	/* (4) Check parameters
	---------------------------------------------------------*/
	var paramError err.Error = err.Success
	parameters := make(map[string]interface{})
	for name, param := range method.Parameters {

		/* (1) Extract value */
		p, isset := request.Data.Set[name]

		/* (2) Required & missing */
		if !isset && !*param.Optional {
			paramError = err.MissingParam
			paramError.BindArgument(name)
			break
		}

		/* (3) Optional & missing: set default value */
		if !isset {
			p = &requestParameter{
				Parsed: true,
				File:   param.Type == "FILE",
				Value:  nil,
			}
			if param.Default != nil {
				p.Value = *param.Default
			}
		}

		/* (4) Parse parameter if not file */
		if !p.Parsed && !p.File {
			p.Value = parseHttpData(p.Value)
		}

		/* (4) Fail on unexpected multipart file */
		waitFile, gotFile := param.Type == "FILE", p.File
		if gotFile && !waitFile || !gotFile && waitFile {
			paramError = err.InvalidParam
			paramError.BindArgument(name)
			paramError.BindArgument("FILE")
			break
		}

		/* (5) Do not check if file */
		if gotFile {
			parameters[name] = p.Value
			continue
		}

		/* (6) Check type */
		if s.Checker.Run(param.Type, p.Value) != nil {

			paramError = err.InvalidParam
			paramError.BindArgument(name)
			paramError.BindArgument(param.Type)
			paramError.BindArgument(p.Value)
			break

		}

		parameters[name] = p.Value

	}

	// Fail if argument check failed
	if paramError.Code != err.Success.Code {
		Json, _ := paramError.MarshalJSON()
		res.Header().Add("Content-Type", "application/json")
		res.Write(Json)
		log.Printf("[err] %s\n", paramError.Reason)
		return
	}

	/* (5) Load controller
	---------------------------------------------------------*/
	callable, err := request.loadController(req.Method)
	if err != nil {
		log.Printf("[err] %s\n", err)
		return
	}
	fmt.Printf("OK\nplugin: '%si.so'\n", strings.Join(request.ControllerUri, "/"))
	for name, value := range parameters {
		fmt.Printf("  $%s = %v\n", name, value)
	}

	/* (6) Execute and get response
	---------------------------------------------------------*/
	out, _ := callable(parameters)
	fmt.Printf("-- OUT --\n")
	for name, value := range out {
		fmt.Printf("  $%s = %v\n", name, value)
	}
	return
}

func (s *Server) findController(req *Request) *config.Controller {
	/* (1) Init browsing cursors */
	ctl := s.config
	uriIndex := 0

	/* (2) Browse while there is uri parts */
	for uriIndex < len(req.Uri) {
		uri := req.Uri[uriIndex]

		child, hasKey := ctl.Children[uri]

		// stop if no matchind child
		if !hasKey {
			break
		}

		req.ControllerUri = append(req.ControllerUri, uri)
		ctl = child
		uriIndex++

	}

	/* (3) Extract & store URI params */
	req.Data.fillUrl(req.Uri[uriIndex:])

	/* (4) Return controller */
	return ctl

}

func (s *Server) getMethod(controller *config.Controller, method string) *config.Method {

	/* (1) Unavailable method */
	if !config.IsMethodAvailable(method) {
		return nil
	}

	/* (2) Extract method cursor */
	var foundMethod = controller.Method(method)

	/* (3) Return method | nil on error */
	return foundMethod

}
