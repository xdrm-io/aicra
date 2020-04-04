package api

import (
	"net/http"
	"strings"
)

// Request represents an API request i.e. HTTP
type Request struct {
	// corresponds to the list of uri components
	//  featured in the request URI
	URI []string

	// Scope from the configuration file of the current service
	Scope [][]string

	// original HTTP request
	Request *http.Request

	// input parameters
	Param RequestParam
}

// NewRequest builds an interface request from a http.Request
func NewRequest(req *http.Request) *Request {
	uri := normaliseURI(req.URL.Path)
	uriparts := strings.Split(uri, "/")

	return &Request{
		URI:     uriparts,
		Scope:   nil,
		Request: req,
		Param:   make(RequestParam),
	}
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
