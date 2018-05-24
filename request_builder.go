package gfw

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// buildRequest builds an interface request
// from a http.Request
func buildRequest(req *http.Request) (*Request, error) {

	/* (1) Init request */
	uri := NormaliseUri(req.URL.Path)
	inst := &Request{
		Uri:      strings.Split(uri, "/"),
		GetData:  FetchGetData(req),
		FormData: FetchFormData(req),
		UrlData:  make([]interface{}, 0),
		Data:     make(map[string]interface{}, 0),
	}
	inst.ControllerUri = make([]string, 0, len(inst.Uri))

	/* (2) Fill 'Data' with GET data */
	for name, data := range inst.GetData {
		// prevent injections
		if isParameterNameInjection(name) {
			log.Printf("get.name_injection:  '%s'\n", name)
			delete(inst.GetData, name)
			continue
		}

		// add into data
		inst.Data[fmt.Sprintf("GET@%s", name)] = data
	}

	/* (3) Fill 'Data' with POST data */
	for name, data := range inst.FormData {

		// prevent injections
		if isParameterNameInjection(name) {
			log.Printf("post.name_injection: '%s'\n", name)
			delete(inst.FormData, name)
			continue
		}

		// add into data
		inst.Data[name] = data
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
