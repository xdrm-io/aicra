package gfw

import (
	"fmt"
	"git.xdrm.io/gfw/internal/config"
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

	/* (4) Check arguments
	---------------------------------------------------------*/
	var paramError Err = ErrSuccess
	for name, param := range method.Parameters {
		fmt.Printf("- %s: %v | '%v'\n", name, *param.Optional, *param.Rename)

		/* (1) Extract value */
		p, isset := request.Data.Set[name]

		/* (2) OPTIONAL ? */
		if !isset {

			// fail if required
			if !*param.Optional {
				paramError = ErrMissingParam
				paramError.BindArgument(name)
				break

				// error if default param is nil
			} else if param.Default == nil {
				paramError = ErrInvalidDefaultParam
				paramError.BindArgument(name)
				break

				// set default p if optional
			} else {
				p = &RequestParameter{
					Parsed: true,
					Value:  *param.Default,
				}
			}

		}

		/* (3) Check type */
		isValid := s.Checker.Run(param.Type, p.Value)
		if isValid != nil {
			paramError = ErrInvalidParam
			paramError.BindArgument(name)
			paramError.BindArgument(param.Type)
			paramError.BindArgument(p.Value)
			break
		}

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
	return
}
