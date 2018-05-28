package gfw

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// buildRequest builds an interface request
// from a http.Request
func buildRequest(req *http.Request) (*Request, error) {

	/* (1) Init request */
	uri := NormaliseUri(req.URL.Path)
	rawpost := FetchFormData(req)
	rawget := FetchGetData(req)
	inst := &Request{
		Uri:      strings.Split(uri, "/"),
		GetData:  make(map[string]interface{}, 0),
		FormData: make(map[string]interface{}, 0),
		UrlData:  make([]interface{}, 0),
		Data:     make(map[string]interface{}, 0),
	}
	inst.ControllerUri = make([]string, 0, len(inst.Uri))

	/* (2) Fill 'Data' with GET data */
	for name, rawdata := range rawget {

		// 1. Parse arguments
		data := parseHttpData(rawdata)

		if data == nil {
			continue
		}

		// 2. prevent injections
		if isParameterNameInjection(name) {
			log.Printf("get.name_injection:  '%s'\n", name)
			delete(inst.GetData, name)
			continue
		}

		// 3. add into data
		inst.GetData[name] = data
		inst.Data[fmt.Sprintf("GET@%s", name)] = data
	}

	/* (3) Fill 'Data' with POST data */
	for name, rawdata := range rawpost {

		// 1. Parse arguments
		data := parseHttpData(rawdata)

		if data == nil {
			continue
		}

		// 2. prevent injections
		if isParameterNameInjection(name) {
			log.Printf("post.name_injection: '%s'\n", name)
			delete(inst.FormData, name)
			continue
		}

		// 3. add into data
		inst.Data[name] = data
		inst.FormData[name] = data
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

// FetchGetData extracts the GET data
// from an HTTP request
func FetchGetData(req *http.Request) map[string]interface{} {

	res := make(map[string]interface{})

	for name, value := range req.URL.Query() {
		res[name] = value
	}

	return res

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

// isParameterNameInjection returns whether there is
// a parameter name injection:
// - inferred GET parameters
// - inferred URL parameters
func isParameterNameInjection(pName string) bool {
	return strings.HasPrefix(pName, "GET@") || strings.HasPrefix(pName, "URL#")
}

// parseHttpData parses http GET/POST data
// - []string of 1 element : return json of element 0
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
	return nil

}
