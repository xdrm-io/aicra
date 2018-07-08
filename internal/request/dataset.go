package request

import (
	"encoding/json"
	"fmt"
	"git.xdrm.io/go/aicra/internal/multipart"
	"log"
	"net/http"
	"strings"
)

func NewDataset() *DataSet {
	return &DataSet{
		URI:  make([]*Parameter, 0),
		Get:  make(map[string]*Parameter),
		Form: make(map[string]*Parameter),
		Set:  make(map[string]*Parameter),
	}
}

// Build builds a 'DataSet' from an http request
func (i *DataSet) Build(req *http.Request) {

	/* (1) GET (query) data */
	i.fetchGet(req)

	/* (2) We are done if GET method */
	if req.Method == "GET" {
		return
	}

	/* (3) POST (body) data */
	i.fetchForm(req)

}

// SetURI stores URL data and fills 'Set'
// with creating pointers inside 'Url'
func (i *DataSet) SetURI(data []string) {

	for index, value := range data {

		// create set index
		setindex := fmt.Sprintf("URL#%d", index)

		// store value in 'Set'
		i.Set[setindex] = &Parameter{
			Parsed: false,
			Value:  value,
		}

		// create link in 'Url'
		i.URI = append(i.URI, i.Set[setindex])

	}

}

// fetchGet stores data from the QUERY (in url parameters)
func (i *DataSet) fetchGet(req *http.Request) {

	for name, value := range req.URL.Query() {

		// prevent injections
		if nameInjection(name) {
			log.Printf("get.injection: '%s'\n", name)
			continue
		}

		// create set index
		setindex := fmt.Sprintf("GET@%s", name)

		// store value in 'Set'
		i.Set[setindex] = &Parameter{
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
func (i *DataSet) fetchForm(req *http.Request) {

	contentType := req.Header.Get("Content-Type")

	// parse json
	if strings.HasPrefix(contentType, "application/json") {
		i.parseJSON(req)
		return
	}

	// parse urlencoded
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		i.parseUrlencoded(req)
		return
	}

	// parse multipart
	if strings.HasPrefix(contentType, "multipart/form-data; boundary=") {
		i.parseMultipart(req)
		return
	}

	// if unknown type store nothing
}

// parseJSON parses JSON from the request body inside 'Form'
// and 'Set'
func (i *DataSet) parseJSON(req *http.Request) {

	parsed := make(map[string]interface{}, 0)

	decoder := json.NewDecoder(req.Body)

	// if parse error: do nothing
	if err := decoder.Decode(&parsed); err != nil {
		return
	}

	// else store values 'parsed' values
	for name, value := range parsed {

		// prevent injections
		if nameInjection(name) {
			log.Printf("post.injection: '%s'\n", name)
			continue
		}

		// store value in 'Set'
		i.Set[name] = &Parameter{
			Parsed: true,
			Value:  value,
		}

		// create link in 'Form'
		i.Form[name] = i.Set[name]

	}

}

// parseUrlencoded parses urlencoded from the request body inside 'Form'
// and 'Set'
func (i *DataSet) parseUrlencoded(req *http.Request) {

	// use http.Request interface
	req.ParseForm()

	for name, value := range req.PostForm {

		// prevent injections
		if nameInjection(name) {
			log.Printf("post.injection: '%s'\n", name)
			continue
		}

		// store value in 'Set'
		i.Set[name] = &Parameter{
			Parsed: false,
			Value:  value,
		}

		// create link in 'Form'
		i.Form[name] = i.Set[name]
	}

}

// parseMultipart parses multi-part from the request body inside 'Form'
// and 'Set'
func (i *DataSet) parseMultipart(req *http.Request) {

	/* (1) Create reader */
	mpr := multipart.CreateReader(req)

	/* (2) Parse multipart */
	mpr.Parse()

	/* (3) Store data into 'Form' and 'Set */
	for name, component := range mpr.Components {

		// prevent injections
		if nameInjection(name) {
			log.Printf("post.injection: '%s'\n", name)
			continue
		}

		// store value in 'Set'
		i.Set[name] = &Parameter{
			Parsed: false,
			File:   component.File,
			Value:  component.Data,
		}

		// create link in 'Form'
		i.Form[name] = i.Set[name]

	}

	return

}
