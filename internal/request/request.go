package request

import (
	"git.xdrm.io/go/aicra/driver"
	"git.xdrm.io/go/aicra/err"
	"git.xdrm.io/go/aicra/response"
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

// RunController tries to load a controller from its uri
// checks for its given method ('Get', 'Post', 'Put', or 'Delete')
func (i *Request) RunController(_method string, _driver driver.Driver) (func(response.Arguments) response.Response, err.Error) {

	return _driver.RunController(i.Path, _method)

}
