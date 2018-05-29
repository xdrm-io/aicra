package gfw

import (
	"encoding/json"
	"fmt"
	"git.xdrm.io/gfw/internal/multipart"
	"log"
	"net/http"
	"strings"
)

// buildRequestDataFromRequest builds a 'RequestData'
// from an http request
func buildRequestDataFromRequest(req *http.Request) *RequestData {

	i := &RequestData{
		Url:  make([]*RequestParameter, 0),
		Get:  make(map[string]*RequestParameter),
		Form: make(map[string]*RequestParameter),
		Set:  make(map[string]*RequestParameter),
	}

	// GET (query) data
	i.fetchGet(req)

	// no Form if GET
	if req.Method == "GET" {
		return i
	}

	// POST (body) data
	i.fetchForm(req)

	return i

}

// bindUrl stores URL data and fills 'Set'
// with creating pointers inside 'Url'
func (i *RequestData) fillUrl(data []string) {

	for index, value := range data {

		// create set index
		setindex := fmt.Sprintf("URL#%d", index)

		// store value in 'Set'
		i.Set[setindex] = &RequestParameter{
			Parsed: false,
			Value:  value,
		}

		// create link in 'Url'
		i.Url = append(i.Url, i.Set[setindex])

	}

}

// fetchGet stores data from the QUERY (in url parameters)
func (i *RequestData) fetchGet(req *http.Request) {

	for name, value := range req.URL.Query() {

		// prevent injections
		if isParameterNameInjection(name) {
			log.Printf("get.injection: '%s'\n", name)
			continue
		}

		// create set index
		setindex := fmt.Sprintf("GET@%s", name)

		// store value in 'Set'
		i.Set[setindex] = &RequestParameter{
			Parsed: false,
			Value:  value,
		}

		// create link in 'Get'
		i.Get[name] = i.Set[setindex]

	}

}

// fetchForm stores FORM data
//
// - parse 'form-data' if not supported (not POST requests)
// - parse 'x-www-form-urlencoded'
// - parse 'application/json'
func (i *RequestData) fetchForm(req *http.Request) {

	contentType := req.Header.Get("Content-Type")

	// parse json
	if strings.HasPrefix(contentType, "application/json") {
		i.parseJsonForm(req)
		return
	}

	// parse urlencoded
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		i.parseUrlencodedForm(req)
		return
	}

	// parse multipart
	if strings.HasPrefix(contentType, "multipart/form-data; boundary=") {
		i.parseMultipartForm(req)
		return
	}

	// if unknown type store nothing
}

// parseJsonForm parses JSON from the request body inside 'Form'
// and 'Set'
func (i *RequestData) parseJsonForm(req *http.Request) {

	parsed := make(map[string]interface{}, 0)

	decoder := json.NewDecoder(req.Body)

	// if parse error: do nothing
	if err := decoder.Decode(&parsed); err != nil {
		return
	}

	// else store values 'parsed' values
	for name, value := range parsed {

		// prevent injections
		if isParameterNameInjection(name) {
			log.Printf("post.injection: '%s'\n", name)
			continue
		}

		// store value in 'Set'
		i.Set[name] = &RequestParameter{
			Parsed: true,
			Value:  value,
		}

		// create link in 'Form'
		i.Form[name] = i.Set[name]

	}

}

// parseUrlencodedForm parses urlencoded from the request body inside 'Form'
// and 'Set'
func (i *RequestData) parseUrlencodedForm(req *http.Request) {

	// use http.Request interface
	req.ParseForm()

	for name, value := range req.PostForm {

		// prevent injections
		if isParameterNameInjection(name) {
			log.Printf("post.injection: '%s'\n", name)
			continue
		}

		// store value in 'Set'
		i.Set[name] = &RequestParameter{
			Parsed: false,
			Value:  value,
		}

		// create link in 'Form'
		i.Form[name] = i.Set[name]
	}

}

// parseMultipartForm parses multi-part from the request body inside 'Form'
// and 'Set'
func (i *RequestData) parseMultipartForm(req *http.Request) {

	/* (1) Create reader */
	mpr := multipart.CreateReader(req)

	/* (2) Parse multipart */
	mpr.Parse()

	/* (3) Store data into 'Form' and 'Set */
	for name, component := range mpr.Components {

		// prevent injections
		if isParameterNameInjection(name) {
			log.Printf("post.injection: '%s'\n", name)
			continue
		}

		// store value in 'Set'
		i.Set[name] = &RequestParameter{
			Parsed: false,
			File:   component.File,
			Value:  component.Data,
		}

		// create link in 'Form'
		i.Form[name] = i.Set[name]

	}

	return

}

// isParameterNameInjection returns whether there is
// a parameter name injection:
// - inferred GET parameters
// - inferred URL parameters
func isParameterNameInjection(pName string) bool {
	return strings.HasPrefix(pName, "GET@") || strings.HasPrefix(pName, "URL#")
}
