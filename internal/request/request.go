package request

import (
	"net/http"
	"strings"
)

// FromHTTP builds an interface request from a http.Request
func FromHTTP(req *http.Request) (*Request, error) {

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
