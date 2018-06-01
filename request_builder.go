package gfw

import (
	"encoding/json"
	"fmt"
	"git.xdrm.io/xdrm-brackets/gfw/err"
	"log"
	"net/http"
	"plugin"
	"reflect"
	"strings"
	"time"
)

// buildRequest builds an interface request
// from a http.Request
func buildRequest(req *http.Request) (*Request, error) {

	/* (1) Get useful data */
	uri := NormaliseUri(req.URL.Path)
	uriparts := strings.Split(uri, "/")

	/* (2) Init request */
	inst := &Request{
		Uri:           uriparts,
		ControllerUri: make([]string, 0, len(uriparts)),
		Data:          buildRequestDataFromRequest(req),
	}

	return inst, nil
}

// NormaliseUri removes the trailing '/' to always
// have the same Uri format for later processing
func NormaliseUri(uri string) string {

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

// parseHttpData parses http GET/POST data
// - []string
//		- size = 1 : return json of first element
//		- size > 1 : return array of json elements
// - string : return json if valid, else return raw string
func parseHttpData(data interface{}) interface{} {
	dtype := reflect.TypeOf(data)
	dvalue := reflect.ValueOf(data)

	switch dtype.Kind() {

	/* (1) []string -> recursive */
	case reflect.Slice:

		// 1. Return nothing if empty
		if dvalue.Len() == 0 {
			return nil
		}

		// 2. only return first element if alone
		if dvalue.Len() == 1 {

			element := dvalue.Index(0)
			if element.Kind() != reflect.String {
				return nil
			}
			return parseHttpData(element.String())

			// 3. Return all elements if more than 1
		} else {

			result := make([]interface{}, dvalue.Len())

			for i, l := 0, dvalue.Len(); i < l; i++ {
				element := dvalue.Index(i)

				// ignore non-string
				if element.Kind() != reflect.String {
					continue
				}

				result[i] = parseHttpData(element.String())
			}
			return result

		}

	/* (2) string -> parse */
	case reflect.String:

		// build json wrapper
		wrapper := fmt.Sprintf("{\"wrapped\":%s}", dvalue.String())

		// try to parse as json
		var result interface{}
		err := json.Unmarshal([]byte(wrapper), &result)

		// return if success
		if err == nil {

			mapval, ok := result.(map[string]interface{})
			if !ok {
				return dvalue.String()
			}

			wrapped, ok := mapval["wrapped"]
			if !ok {
				return dvalue.String()
			}

			return wrapped
		}

		// else return as string
		return dvalue.String()

	}

	/* (3) NIL if unknown type */
	return dvalue

}

// loadController tries to load a controller from its uri
// checks for its given method ('Get', 'Post', 'Put', or 'Delete')
func (i *Request) loadController(method string) (func(map[string]interface{}) (map[string]interface{}, err.Error), error) {

	/* (1) Build controller path */
	path := fmt.Sprintf("%si.so", i.ControllerUri)

	/* (2) Format url */
	tmp := []byte(strings.ToLower(method))
	tmp[0] = tmp[0] - ('a' - 'A')
	method = string(tmp)

	fmt.Printf("method is '%s'\n", method)
	return nil, nil

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
	callable, validSignature := m.(func(map[string]interface{}) (map[string]interface{}, err.Error))
	if !validSignature {
		return nil, fmt.Errorf("Invalid signature for method %s", method)
	}

	return callable, nil

}
