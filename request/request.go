package request

import (
	"encoding/json"
	"fmt"
	"git.xdrm.io/go/aicra/implement"
	"log"
	"net/http"
	"plugin"
	"strings"
	"time"
)

// Build builds an interface request from a http.Request
func Build(req *http.Request) (*Request, error) {

	/* (1) Get useful data */
	uri := normaliseUri(req.URL.Path)
	uriparts := strings.Split(uri, "/")

	/* (2) Init request */
	inst := &Request{
		Uri:  uriparts,
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
func (i *Request) LoadController(method string) (func(implement.Arguments, *implement.Response) implement.Response, error) {

	/* (1) Build controller path */
	path := fmt.Sprintf("./controllers/%si.so", strings.Join(i.Path, "/"))

	/* (2) Format url */
	tmp := []byte(strings.ToLower(method))
	tmp[0] = tmp[0] - ('a' - 'A')
	method = string(tmp)

	/* (2) Try to load plugin */
	p, err2 := plugin.Open(path)
	if err2 != nil {
		return nil, err2
	}

	/* (3) Try to extract method */
	m, err2 := p.Lookup(method)
	if err2 != nil {
		return nil, err2
	}

	/* (4) Check signature */
	callable, validSignature := m.(func(implement.Arguments, *implement.Response) implement.Response)
	if !validSignature {
		return nil, fmt.Errorf("Invalid signature for method %s", method)
	}

	return callable, nil

}
