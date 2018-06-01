package gfw

import (
	"fmt"
	"git.xdrm.io/xdrm-brackets/gfw/config"
	"git.xdrm.io/xdrm-brackets/gfw/err"
	"git.xdrm.io/xdrm-brackets/gfw/request"
	"log"
	"net/http"
	"strings"
)

func (s *Server) route(res http.ResponseWriter, httpReq *http.Request) {

	/* (1) Build request
	---------------------------------------------------------*/
	/* (1) Try to build request */
	req, err2 := request.Build(httpReq)
	if err2 != nil {
		log.Fatal(err2)
	}

	/* (2) Find a controller
	---------------------------------------------------------*/
	controller := s.findController(req)

	/* (3) Check method
	---------------------------------------------------------*/
	var method *config.Method
	if method = controller.Method(httpReq.Method); method == nil {
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
		p, isset := req.Data.Set[name]

		/* (2) Required & missing */
		if !isset && !*param.Optional {
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
		}

		/* (4) Parse parameter if not file */
		if !p.File {
			p.Parse()
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
	callable, err := req.LoadController(httpReq.Method)
	if err != nil {
		log.Printf("[err] %s\n", err)
		return
	}
	fmt.Printf("OK\nplugin: '%si.so'\n", strings.Join(req.Path, "/"))
	for name, value := range parameters {
		fmt.Printf("  $%s = %v\n", name, value)
	}

	/* (6) Execute and get response
	---------------------------------------------------------*/
	resp := callable(parameters)
	if resp != nil {
		fmt.Printf("-- OUT --\n")
		for name, value := range resp.Dump() {
			fmt.Printf("  $%s = %v\n", name, value)
		}
		eJSON, _ := resp.Err.MarshalJSON()
		fmt.Printf("-- ERR --\n%s\n", eJSON)
	}
	return
}

func (s *Server) findController(req *request.Request) *config.Controller {

	/* (1) Try to browse by URI */
	pathi, ctl := s.config.Browse(req.Uri)

	/* (2) Set controller uri */
	req.Path = make([]string, 0, pathi)
	req.Path = append(req.Path, req.Uri[:pathi]...)

	/* (3) Extract & store URI params */
	req.Data.SetUri(req.Uri[pathi:])

	/* (4) Return controller */
	return ctl

}
