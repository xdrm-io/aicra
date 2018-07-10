package aicra

import (
	"encoding/json"
	"git.xdrm.io/go/aicra/err"
	"git.xdrm.io/go/aicra/internal/apirequest"
	"git.xdrm.io/go/aicra/internal/controller"
	"git.xdrm.io/go/aicra/response"
	"log"
	"net/http"
)

func (s *Server) findController(req *apirequest.Request) *controller.Controller {

	/* (1) Try to browse by URI */
	pathi, ctl := s.controller.Browse(req.URI)

	/* (2) Set controller uri */
	req.Path = make([]string, 0, pathi)
	req.Path = append(req.Path, req.URI[:pathi]...)

	/* (3) Extract & store URI params */
	req.Data.SetURI(req.URI[pathi:])

	/* (4) Return controller */
	return ctl

}

// Redirects to another location (http protocol)
func httpRedirect(r http.ResponseWriter, loc string) {
	r.Header().Add("Location", loc)
	r.WriteHeader(308) // permanent redirect
}

// Prints an HTTP response
func httpPrint(r http.ResponseWriter, res response.Response) {
	// get response data
	formattedResponse := res.Dump()

	// add error fields
	formattedResponse["error"] = res.Err.Code
	formattedResponse["reason"] = res.Err.Reason

	// add arguments if any
	if res.Err.Arguments != nil && len(res.Err.Arguments) > 0 {
		formattedResponse["args"] = res.Err.Arguments
	}

	// write this json
	jsonResponse, _ := json.Marshal(formattedResponse)
	r.Header().Add("Content-Type", "application/json")
	r.Write(jsonResponse)
}

// Prints an error as HTTP response
func httpError(r http.ResponseWriter, e err.Error) {
	JSON, _ := e.MarshalJSON()
	r.Header().Add("Content-Type", "application/json")
	r.Write(JSON)
	log.Printf("[http.fail] %s\n", e.Reason)
}
