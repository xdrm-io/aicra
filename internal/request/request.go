package request

import (
	"net/http"
	"strings"
)

// Request represents an API request i.e. HTTP
type Request struct {
	// corresponds to the list of uri components
	//  featuring in the request URI
	URI []string

	// controller path (portion of 'Uri')
	Path []string

	// contains all data from URL, GET, and FORM
	Data *DataSet
}

// New builds an interface request from a http.Request
func New(req *http.Request) (*Request, error) {

	/* (1) Get useful data */
	uri := normaliseURI(req.URL.Path)
	uriparts := strings.Split(uri, "/")

	/* (2) Init request */
	inst := &Request{
		URI:  uriparts,
		Path: make([]string, 0, len(uriparts)),
		Data: NewDataset(),
	}

	/* (3) Build dataset */
	inst.Data.Build(req)

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
