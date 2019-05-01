package api

import (
	"net/http"
	"strings"
)

// RequestParam defines input parameters of an api request
type RequestParam map[string]interface{}

// Request represents an API request i.e. HTTP
type Request struct {
	// corresponds to the list of uri components
	//  featured in the request URI
	URI []string

	// original HTTP request
	Request *http.Request

	// input parameters
	Param RequestParam
}

// NewRequest builds an interface request from a http.Request
func NewRequest(req *http.Request) (*Request, error) {

	// 1. get useful data
	uri := normaliseURI(req.URL.Path)
	uriparts := strings.Split(uri, "/")

	// 3. Init request
	inst := &Request{
		URI:     uriparts,
		Request: req,
		Param:   make(RequestParam),
	}

	return inst, nil
}

// normaliseURI removes the trailing '/' to always
// have the same Uri format for later processing
func normaliseURI(uri string) string {

	if len(uri) < 1 {
		return uri
	}

	if uri[0] == '/' {
		uri = uri[1:]
	}

	if len(uri) > 1 && uri[len(uri)-1] == '/' {
		uri = uri[0 : len(uri)-1]
	}

	return uri
}
