package request

import (
	"encoding/json"
	"fmt"
	"git.xdrm.io/go/aicra/driver"
	"git.xdrm.io/go/aicra/err"
	"git.xdrm.io/go/aicra/response"
	"log"
	"net/http"
	"strings"
	"time"
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

// FetchFormData extracts FORM data
//
// - parse 'form-data' if not supported (not POST requests)
// - parse 'x-www-form-urlencoded'
// - parse 'application/json'
func FetchFormData(req *http.Request) map[string]interface{} {

	res := make(map[string]interface{})

	// Abort if GET request
	if req.Method == "GET" {
		return res
	}

	ct := req.Header.Get("Content-Type")

	if strings.HasPrefix(ct, "application/json") {

		receiver := make(map[string]interface{}, 0)

		// 1. Init JSON reader
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&receiver); err != nil {
			log.Printf("[parse.json] %s\n", err)
			return res
		}

		// 2. Return result
		return receiver

	} else if strings.HasPrefix(ct, "application/x-www-form-urlencoded") {

		// 1. Parse url encoded data
		req.ParseForm()

		// 2. Extract values
		for name, value := range req.PostForm {
			res[name] = value
		}

	} else { // form-data or anything

		startn := time.Now().UnixNano()
		// 1. Parse form-data
		if err := req.ParseMultipartForm(req.ContentLength + 1); err != nil {
			log.Printf("[read.multipart] %s\n", err)
			return res
		}

		// 2. Extract values
		for name, value := range req.PostForm {
			res[name] = value
		}
		fmt.Printf("* %.3f us\n", float64(time.Now().UnixNano()-startn)/1e3)

	}

	return res
}

// LoadController tries to load a controller from its uri
// checks for its given method ('Get', 'Post', 'Put', or 'Delete')
func (i *Request) LoadController(_method string, _driver driver.Driver) (func(response.Arguments) response.Response, err.Error) {

	return _driver.Load(i.Path, _method)

}
