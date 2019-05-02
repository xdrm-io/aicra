package reqdata

import (
	"encoding/json"
	"fmt"
	"log"

	"git.xdrm.io/go/aicra/internal/multipart"

	"net/http"
	"strings"
)

// Store represents all data that can be caught:
// - URI (guessed from the URI by removing the service path)
// - GET (default url data)
// - POST (from json, form-data, url-encoded)
type Store struct {

	// ordered values from the URI
	//  catches all after the service path
	//
	// points to Store.Data
	URI []*Parameter

	// uri parameters following the QUERY format
	//
	// points to Store.Data
	Get map[string]*Parameter

	// form data depending on the Content-Type:
	//  'application/json'                  => key-value pair is parsed as json into the map
	//  'application/x-www-form-urlencoded' => standard parameters as QUERY parameters
	//  'multipart/form-data'               => parse form-data format
	//
	// points to Store.Data
	Form map[string]*Parameter

	// contains URL+GET+FORM data with prefixes:
	// - FORM: no prefix
	// - URL:  'URL#' followed by the index in Uri
	// - GET:  'GET@' followed by the key in GET
	Set map[string]*Parameter
}

// New creates a new store from an http request.
// URI params is required because it only takes into account after service path
// we do not know in this scope.
func New(uriParams []string, req *http.Request) *Store {
	ds := &Store{
		URI:  make([]*Parameter, 0),
		Get:  make(map[string]*Parameter),
		Form: make(map[string]*Parameter),
		Set:  make(map[string]*Parameter),
	}

	// 1. set URI parameters
	ds.setURIParams(uriParams)

	// 2. GET (query) data
	ds.readQuery(req)

	// 3. We are done if GET method
	if req.Method == http.MethodGet {
		return ds
	}

	// 4. POST (body) data
	ds.readForm(req)

	return ds
}

// setURIParameters fills 'Set' with creating pointers inside 'Url'
func (i *Store) setURIParams(orderedUParams []string) {

	for index, value := range orderedUParams {

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

// readQuery stores data from the QUERY (in url parameters)
func (i *Store) readQuery(req *http.Request) {

	for name, value := range req.URL.Query() {

		// prevent invalid names
		if !isNameValid(name) {
			log.Printf("invalid variable name: '%s'\n", name)
			continue
		}

		// prevent injections
		if hasNameInjection(name) {
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

// readForm stores FORM data
//
// - parse 'form-data' if not supported (not POST requests)
// - parse 'x-www-form-urlencoded'
// - parse 'application/json'
func (i *Store) readForm(req *http.Request) {

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
func (i *Store) parseJSON(req *http.Request) {

	parsed := make(map[string]interface{}, 0)

	decoder := json.NewDecoder(req.Body)

	// if parse error: do nothing
	if err := decoder.Decode(&parsed); err != nil {
		log.Printf("json.parse() %s\n", err)
		return
	}

	// else store values 'parsed' values
	for name, value := range parsed {

		// prevent invalid names
		if !isNameValid(name) {
			log.Printf("invalid variable name: '%s'\n", name)
			continue
		}

		// prevent injections
		if hasNameInjection(name) {
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
func (i *Store) parseUrlencoded(req *http.Request) {

	// use http.Request interface
	if err := req.ParseForm(); err != nil {
		log.Printf("urlencoded.parse() %s\n", err)
		return
	}

	for name, value := range req.PostForm {

		// prevent invalid names
		if !isNameValid(name) {
			log.Printf("invalid variable name: '%s'\n", name)
			continue
		}

		// prevent injections
		if hasNameInjection(name) {
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
func (i *Store) parseMultipart(req *http.Request) {

	/* (1) Create reader */
	boundary := req.Header.Get("Content-Type")[len("multipart/form-data; boundary="):]
	mpr, err := multipart.NewReader(req.Body, boundary)
	if err != nil {
		return
	}

	/* (2) Parse multipart */
	if err = mpr.Parse(); err != nil {
		log.Printf("multipart.parse() %s\n", err)
		return
	}

	/* (3) Store data into 'Form' and 'Set */
	for name, data := range mpr.Data {

		// prevent invalid names
		if !isNameValid(name) {
			log.Printf("invalid variable name: '%s'\n", name)
			continue
		}

		// prevent injections
		if hasNameInjection(name) {
			log.Printf("post.injection: '%s'\n", name)
			continue
		}

		// store value in 'Set'
		i.Set[name] = &Parameter{
			Parsed: false,
			File:   len(data.GetHeader("filename")) > 0,
			Value:  string(data.Data),
		}

		// create link in 'Form'
		i.Form[name] = i.Set[name]

	}

	return

}

// hasNameInjection returns whether there is
// a parameter name injection:
// - inferred GET parameters
// - inferred URL parameters
func hasNameInjection(pName string) bool {
	return strings.HasPrefix(pName, "GET@") || strings.HasPrefix(pName, "URL#")
}

// isNameValid returns whether a parameter name (without the GET@ or URL# prefix) is valid
// if fails if the name begins/ends with underscores
func isNameValid(pName string) bool {
	return strings.Trim(pName, "_") == pName
}
