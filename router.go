package gfw

import (
	"encoding/json"
	"git.xdrm.io/go/xb-api/config"
	"git.xdrm.io/go/xb-api/err"
	"git.xdrm.io/go/xb-api/implement"
	"git.xdrm.io/go/xb-api/request"
	"log"
	"net/http"
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

	/* (6) Execute and get response
	---------------------------------------------------------*/
	/* (1) Execute */
	responseBarebone := implement.NewResponse()
	response := callable(parameters, responseBarebone)

	/* (2) Build JSON response */
	formattedResponse := response.Dump()
	formattedResponse["error"] = response.Err.Code
	formattedResponse["reason"] = response.Err.Reason
	if response.Err.Arguments != nil && len(response.Err.Arguments) > 0 {
		formattedResponse["args"] = response.Err.Arguments
	}
	jsonResponse, _ := json.Marshal(formattedResponse)

	res.Header().Add("Content-Type", "application/json")
	res.Write(jsonResponse)
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
