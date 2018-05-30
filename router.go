package gfw

import (
	"fmt"
	"git.xdrm.io/xdrm-brackets/gfw/config"
	"log"
	"net/http"
	"strings"
)

func (s *Server) route(res http.ResponseWriter, req *http.Request) {

	/* (1) Build request
	---------------------------------------------------------*/
	/* (1) Try to build request */
	request, err := buildRequest(req)
	if err != nil {
		log.Fatal(req)
	}

	/* (2) Find a controller
	---------------------------------------------------------*/
	/* (1) Init browsing cursors */
	ctl := s.config
	uriIndex := 0

	/* (2) Browse while there is uri parts */
	for uriIndex < len(request.Uri) {
		uri := request.Uri[uriIndex]

		child, hasKey := ctl.Children[uri]

		// stop if no matchind child
		if !hasKey {
			break
		}

		request.ControllerUri = append(request.ControllerUri, uri)
		ctl = child
		uriIndex++

	}

	/* (3) Extract & store URI params */
	request.Data.fillUrl(request.Uri[uriIndex:])

	/* (3) Check method
	---------------------------------------------------------*/
	/* (1) Unavailable method */
	if !config.IsMethodAvailable(req.Method) {

		Json, _ := ErrUnknownMethod.MarshalJSON()
		res.Header().Add("Content-Type", "application/json")
		res.Write(Json)
		log.Printf("[err] %s\n", ErrUnknownMethod.Reason)
		return

	}

	/* (2) Extract method cursor */
	var method = ctl.Method(req.Method)

	/* (3) Unmanaged HTTP method */
	if method == nil { // unknown method
		Json, _ := ErrUnknownMethod.MarshalJSON()
		res.Header().Add("Content-Type", "application/json")
		res.Write(Json)
		log.Printf("[err] %s\n", ErrUnknownMethod.Reason)
		return
	}

	/* (4) Check parameters
	---------------------------------------------------------*/
	var paramError Err = ErrSuccess
	parameters := make(map[string]interface{})
	for name, param := range method.Parameters {

		/* (1) Extract value */
		p, isset := request.Data.Set[name]

		/* (2) Required & missing */
		if !isset && !*param.Optional {
			paramError = ErrMissingParam
			paramError.BindArgument(name)
			break
		}

		/* (3) Optional & missing: set default value */
		if !isset {
			p = &RequestParameter{
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
			paramError = ErrInvalidParam
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

			paramError = ErrInvalidParam
			paramError.BindArgument(name)
			paramError.BindArgument(param.Type)
			paramError.BindArgument(p.Value)
			break

		}

		parameters[name] = p.Value

	}
	fmt.Printf("\n")

	// Fail if argument check failed
	if paramError.Code != ErrSuccess.Code {
		Json, _ := paramError.MarshalJSON()
		res.Header().Add("Content-Type", "application/json")
		res.Write(Json)
		log.Printf("[err] %s\n", paramError.Reason)
		return
	}

	fmt.Printf("OK\nplugin: '%si.so'\n", strings.Join(request.ControllerUri, "/"))
	for name, value := range parameters {
		fmt.Printf("  $%s = %v\n", name, value)
	}
	return
}
