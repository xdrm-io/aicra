package reqdata

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"git.xdrm.io/go/aicra/internal/config"
	"git.xdrm.io/go/aicra/internal/multipart"

	"net/http"
	"strings"
)

// Set represents all data that can be caught:
// - URI (from the URI)
// - GET (default url data)
// - POST (from json, form-data, url-encoded)
//   - 'application/json'                  => key-value pair is parsed as json into the map
//   - 'application/x-www-form-urlencoded' => standard parameters as QUERY parameters
//   - 'multipart/form-data'               => parse form-data format
type Set struct {
	service *config.Service

	// contains URL+GET+FORM data with prefixes:
	// - FORM: no prefix
	// - URL:  '{uri_var}'
	// - GET:  'GET@' followed by the key in GET
	Data map[string]interface{}
}

// New creates a new empty store.
func New(service *config.Service) *Set {
	return &Set{
		service: service,
		Data:    make(map[string]interface{}),
	}
}

// ExtractURI fills 'Set' with creating pointers inside 'Url'
func (i *Set) ExtractURI(req *http.Request) error {
	uriparts := config.SplitURL(req.URL.RequestURI())

	for _, capture := range i.service.Captures {
		// out of range
		if capture.Index > len(uriparts)-1 {
			return fmt.Errorf("%s: %w", capture.Name, ErrMissingURIParameter)
		}
		value := uriparts[capture.Index]

		// should not happen
		if capture.Ref == nil {
			return fmt.Errorf("%s: %w", capture.Name, ErrUnknownType)
		}

		// parse parameter
		parsed := parseParameter(value)

		// check type
		cast, valid := capture.Ref.Validator(parsed)
		if !valid {
			return fmt.Errorf("%s: %w", capture.Name, ErrInvalidType)
		}

		// store cast value in 'Set'
		i.Data[capture.Ref.Rename] = cast
	}

	return nil
}

// ExtractQuery data from the url query parameters
func (i *Set) ExtractQuery(req *http.Request) error {
	query := req.URL.Query()

	for name, param := range i.service.Query {
		value, exist := query[name]

		// fail on missing required
		if !exist && !param.Optional {
			return fmt.Errorf("%s: %w", name, ErrMissingRequiredParam)
		}

		// optional
		if !exist {
			continue
		}

		// parse parameter
		parsed := parseParameter(value)

		// check type
		cast, valid := param.Validator(parsed)
		if !valid {
			return fmt.Errorf("%s: %w", name, ErrInvalidType)
		}

		// store cast value
		i.Data[param.Rename] = cast
	}

	return nil
}

// ExtractForm data from request
//
// - parse 'form-data' if not supported for non-POST requests
// - parse 'x-www-form-urlencoded'
// - parse 'application/json'
func (i *Set) ExtractForm(req *http.Request) error {

	// ignore GET method
	if req.Method == http.MethodGet {
		return nil
	}

	contentType := req.Header.Get("Content-Type")

	// parse json
	if strings.HasPrefix(contentType, "application/json") {
		return i.parseJSON(req)
	}

	// parse urlencoded
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		return i.parseUrlencoded(req)
	}

	// parse multipart
	if strings.HasPrefix(contentType, "multipart/form-data; boundary=") {
		return i.parseMultipart(req)
	}

	// nothing to parse
	return nil
}

// parseJSON parses JSON from the request body inside 'Form'
// and 'Set'
func (i *Set) parseJSON(req *http.Request) error {

	parsed := make(map[string]interface{}, 0)

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&parsed); err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("%s: %w", err, ErrInvalidJSON)
	}

	for name, param := range i.service.Form {
		value, exist := parsed[name]

		// fail on missing required
		if !exist && !param.Optional {
			return fmt.Errorf("%s: %w", name, ErrMissingRequiredParam)
		}

		// optional
		if !exist {
			continue
		}

		// fail on invalid type
		cast, valid := param.Validator(value)
		if !valid {
			return fmt.Errorf("%s: %w", name, ErrInvalidType)
		}

		// store cast value
		i.Data[param.Rename] = cast
	}

	return nil
}

// parseUrlencoded parses urlencoded from the request body inside 'Form'
// and 'Set'
func (i *Set) parseUrlencoded(req *http.Request) error {
	// use http.Request interface
	if err := req.ParseForm(); err != nil {
		return err
	}

	for name, param := range i.service.Form {
		value, exist := req.PostForm[name]

		// fail on missing required
		if !exist && !param.Optional {
			return fmt.Errorf("%s: %w", name, ErrMissingRequiredParam)
		}

		// optional
		if !exist {
			continue
		}

		// parse parameter
		parsed := parseParameter(value)

		// check type
		cast, valid := param.Validator(parsed)
		if !valid {
			return fmt.Errorf("%s: %w", name, ErrInvalidType)
		}

		// store cast value
		i.Data[param.Rename] = cast
	}

	return nil
}

// parseMultipart parses multi-part from the request body inside 'Form'
// and 'Set'
func (i *Set) parseMultipart(req *http.Request) error {

	// 1. create reader
	boundary := req.Header.Get("Content-Type")[len("multipart/form-data; boundary="):]
	mpr, err := multipart.NewReader(req.Body, boundary)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	// 2. parse multipart
	if err = mpr.Parse(); err != nil {
		return fmt.Errorf("%s: %w", err, ErrInvalidMultipart)
	}

	for name, param := range i.service.Form {
		component, exist := mpr.Data[name]

		// fail on missing required
		if !exist && !param.Optional {
			return fmt.Errorf("%s: %w", name, ErrMissingRequiredParam)
		}

		// optional
		if !exist {
			continue
		}

		// parse parameter
		parsed := parseParameter(string(component.Data))

		// fail on invalid type
		cast, valid := param.Validator(parsed)
		if !valid {
			return fmt.Errorf("%s: %w", name, ErrInvalidType)
		}

		// store cast value
		i.Data[param.Rename] = cast
	}

	return nil

}

// parseParameter parses http URI/GET/POST data
// - []string : return array of json elements
// - string   : return json if valid, else return raw string
func parseParameter(data interface{}) interface{} {
	dtype := reflect.TypeOf(data)
	dvalue := reflect.ValueOf(data)

	switch dtype.Kind() {

	/* (1) []string -> recursive */
	case reflect.Slice:

		// 1. ignore empty
		if dvalue.Len() == 0 {
			return data
		}

		// 2. parse each element recursively
		result := make([]interface{}, dvalue.Len())

		for i, l := 0, dvalue.Len(); i < l; i++ {
			element := dvalue.Index(i)
			result[i] = parseParameter(element.Interface())
		}
		return result

	/* (2) string -> parse */
	case reflect.String:

		// build json wrapper
		wrapper := fmt.Sprintf("{\"wrapped\":%s}", dvalue.String())

		// try to parse as json
		var result interface{}
		err := json.Unmarshal([]byte(wrapper), &result)

		// return if success
		if err != nil {
			return dvalue.String()
		}

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

	/* (3) NIL if unknown type */
	return dvalue.Interface()

}
