package aicra

import (
	"encoding/json"
	"log"
	"net/http"

	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/internal/config"
	"git.xdrm.io/go/aicra/internal/reqdata"
)

// extractParameters extracts parameters for the request and checks
// every single one according to configuration options
func (s *Server) extractParameters(store *reqdata.Store, methodParam map[string]*config.Parameter) (map[string]interface{}, api.Error) {

	// init vars
	apiError := api.ErrorSuccess()
	parameters := make(map[string]interface{})

	// for each param of the config
	for name, param := range methodParam {

		/* (1) Extract value */
		p, isset := store.Set[name]

		/* (2) Required & missing */
		if !isset && !param.Optional {
			apiError = api.ErrorMissingParam()
			apiError.Put(name)
			return nil, apiError
		}

		/* (3) Optional & missing: set default value */
		if !isset {
			p = &reqdata.Parameter{
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
			apiError = api.ErrorInvalidParam()
			apiError.Put(param.Rename)
			apiError.Put("FILE")
			return nil, apiError
		}

		/* (6) Do not check if file */
		if gotFile {
			parameters[param.Rename] = p.Value
			continue
		}

		/* (7) Check type */
		if s.Checkers.Run(param.Type, p.Value) != nil {

			apiError = api.ErrorInvalidParam()
			apiError.Put(param.Rename)
			apiError.Put(param.Type)
			apiError.Put(p.Value)
			break

		}

		parameters[param.Rename] = p.Value

	}

	return parameters, apiError
}

// Prints an HTTP response
func httpPrint(r http.ResponseWriter, res *api.Response) {

	// write this json
	jsonResponse, _ := json.Marshal(res)
	r.Header().Add("Content-Type", "application/json")
	r.Write(jsonResponse)
}

// Prints an error as HTTP response
func httpError(r http.ResponseWriter, e api.Error) {
	JSON, _ := json.Marshal(e)
	r.Header().Add("Content-Type", "application/json")
	r.Write(JSON)
	log.Printf("[http.fail] %s\n", e.Reason)
}
