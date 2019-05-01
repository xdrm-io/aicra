package aicra

import (
	"log"

	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/internal/config"
	"git.xdrm.io/go/aicra/internal/reqdata"
)

// extractParameters extracts parameters for the request and checks
// every single one according to configuration options
func (s *Server) extractParameters(store *reqdata.Store, methodParam map[string]*config.Parameter) (map[string]interface{}, api.Error) {

	// init vars
	parameters := make(map[string]interface{})

	// for each param of the config
	for name, param := range methodParam {

		// 1. extract value
		p, isset := store.Set[name]

		// 2. fail if required & missing
		if !isset && !param.Optional {
			apiErr := api.ErrorMissingParam()
			apiErr.SetArguments(name)
			return nil, apiErr
		}

		// 3. optional & missing: set default value
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

		// 4. parse parameter if not file
		if !p.File {
			p.Parse()
		}

		// 5. fail on unexpected multipart file
		waitFile, gotFile := param.Type == "FILE", p.File
		if gotFile && !waitFile || !gotFile && waitFile {
			apiErr := api.ErrorInvalidParam()
			apiErr.SetArguments(param.Rename, "FILE")
			return nil, apiErr
		}

		// 6. do not check if file
		if gotFile {
			parameters[param.Rename] = p.Value
			continue
		}

		// 7. check type
		if s.Checkers.Run(param.Type, p.Value) != nil {
			apiErr := api.ErrorInvalidParam()
			apiErr.SetArguments(param.Rename, param.Type, p.Value)
			return nil, apiErr
		}

		parameters[param.Rename] = p.Value

	}

	return parameters, api.ErrorSuccess()
}

// Prints an error as HTTP response
func logError(res *api.Response) {
	log.Printf("[http.fail] %v\n", res.Err)
}
