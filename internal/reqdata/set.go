package reqdata

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/aicra/internal/multipart"

	"net/http"
	"strings"
)

// T represents all data that can be caught from an http request for a specific
// configuration Service; it features:
// - URI (from the URI)
// - GET (standard url data)
// - POST (from json, form-data, url-encoded)
//   - 'application/json'                  => key-value pair is parsed as json into the map
//   - 'application/x-www-form-urlencoded' => standard parameters as QUERY parameters
//   - 'multipart/form-data'               => parse form-data format
type T struct {
	service *config.Service
	Data    map[string]interface{}
}

// New creates a new empty store.
func New(service *config.Service) *T {
	return &T{
		service: service,
		Data:    map[string]interface{}{},
	}
}

// GetURI parameters
func (i *T) GetURI(req http.Request) error {
	uriparts := config.SplitURL(req.URL.RequestURI())

	for _, capture := range i.service.Captures {
		// out of range
		if capture.Index > len(uriparts)-1 {
			return &Err{field: capture.Ref.Rename, err: ErrMissingURIParameter}
		}
		value := uriparts[capture.Index]

		// should not happen
		if capture.Ref == nil {
			return &Err{field: capture.Ref.Rename, err: ErrUnknownType}
		}

		parsed := parseParameter(value)
		cast, valid := capture.Ref.Validator(parsed)
		if !valid {
			return &Err{field: capture.Ref.Rename, err: ErrInvalidType}
		}
		i.Data[capture.Ref.Rename] = cast
	}

	return nil
}

// GetQuery data from the url query parameters
func (i *T) GetQuery(req http.Request) error {
	query := req.URL.Query()

	for name, param := range i.service.Query {
		values, exist := query[name]

		if !exist {
			if !param.Optional {
				return &Err{field: param.Rename, err: ErrMissingRequiredParam}
			}
			continue
		}

		var parsed interface{}

		// consider slice only if we expect a slice, otherwise, only take the first parameter
		if param.GoType.Kind() == reflect.Slice {
			parsed = parseParameter(values)
		} else {
			// should expect at most 1 value
			if len(values) > 1 {
				return &Err{field: param.Rename, err: ErrInvalidType}
			}
			if len(values) > 0 {
				parsed = parseParameter(values[0])
			}
		}

		cast, valid := param.Validator(parsed)
		if !valid {
			return &Err{field: param.Rename, err: ErrInvalidType}
		}
		i.Data[param.Rename] = cast
	}

	return nil
}

// GetForm parameters the from request
// - parse 'form-data' if not supported for non-POST requests
// - parse 'x-www-form-urlencoded'
// - parse 'application/json'
func (i *T) GetForm(req http.Request) error {
	if req.Method == http.MethodGet {
		return nil
	}

	ct := req.Header.Get("Content-Type")
	switch {
	case strings.HasPrefix(ct, "application/json"):
		err := i.parseJSON(req)
		if err != nil {
			return err
		}

	case strings.HasPrefix(ct, "application/x-www-form-urlencoded"):
		err := i.parseUrlencoded(req)
		if err != nil {
			return err
		}

	case strings.HasPrefix(ct, "multipart/form-data; boundary="):
		err := i.parseMultipart(req)
		if err != nil {
			return err
		}
	}

	// fail on at least 1 mandatory form param when there is no body
	for _, param := range i.service.Form {
		_, exists := i.Data[param.Rename]
		if !exists && !param.Optional {
			return &Err{field: param.Rename, err: ErrMissingRequiredParam}
		}
	}
	return nil
}

// parseJSON parses JSON from the request body inside 'Form'
// and 'Set'
func (i *T) parseJSON(req http.Request) error {
	var parsed map[string]interface{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&parsed)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%s: %w", err, ErrInvalidJSON)
	}

	for name, param := range i.service.Form {
		value, exist := parsed[name]

		if !exist {
			continue
		}

		cast, valid := param.Validator(value)
		if !valid {
			return &Err{field: param.Rename, err: ErrInvalidType}
		}
		i.Data[param.Rename] = cast
	}

	return nil
}

// parseUrlencoded parses urlencoded from the request body inside 'Form'
// and 'Set'
func (i *T) parseUrlencoded(req http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	for name, param := range i.service.Form {
		values, exist := req.PostForm[name]

		if !exist {
			continue
		}

		var parsed interface{}

		// consider slice only if we expect a slice, otherwise, only take the first parameter
		if param.GoType.Kind() == reflect.Slice {
			parsed = parseParameter(values)
		} else if len(values) > 0 {
			// should expect at most 1 value
			if len(values) > 1 {
				return &Err{field: param.Rename, err: ErrInvalidType}
			}
			if len(values) > 0 {
				parsed = parseParameter(values[0])
			}
		}

		cast, valid := param.Validator(parsed)
		if !valid {
			return &Err{field: param.Rename, err: ErrInvalidType}
		}
		i.Data[param.Rename] = cast
	}

	return nil
}

// parseMultipart parses multi-part from the request body inside 'Form'
// and 'Set'
func (i *T) parseMultipart(req http.Request) error {
	boundary := req.Header.Get("Content-Type")[len("multipart/form-data; boundary="):]
	mpr, err := multipart.NewReader(req.Body, boundary)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%s: %w", err, ErrInvalidMultipart)
	}

	err = mpr.Parse()
	if err != nil {
		return fmt.Errorf("%s: %w", err, ErrInvalidMultipart)
	}

	for name, param := range i.service.Form {
		component, exist := mpr.Data[name]

		if !exist {
			continue
		}

		parsed := parseParameter(component.Data)
		cast, valid := param.Validator(parsed)
		if !valid {
			return &Err{field: param.Rename, err: ErrInvalidType}
		}
		i.Data[param.Rename] = cast
	}

	return nil

}

// parseParameter parses http URI/GET/POST data
// - []string : return array of json elements
// - string   : return json if valid, else return raw string
func parseParameter(data interface{}) interface{} {
	rt := reflect.TypeOf(data)
	rv := reflect.ValueOf(data)

	switch rt.Kind() {

	// []string -> recursive
	case reflect.Slice:
		if rv.Len() == 0 {
			return data
		}

		slice := make([]interface{}, rv.Len())
		for i, l := 0, rv.Len(); i < l; i++ {
			element := rv.Index(i)
			slice[i] = parseParameter(element.Interface())
		}
		return slice

	// string -> parse as json
	// keep as string if invalid json
	case reflect.String:
		var cast interface{}
		wrapper := fmt.Sprintf("{\"wrapped\":%s}", rv.String())
		err := json.Unmarshal([]byte(wrapper), &cast)
		if err != nil {
			return rv.String()
		}

		mapval, ok := cast.(map[string]interface{})
		if !ok {
			return rv.String()
		}

		wrapped, ok := mapval["wrapped"]
		if !ok {
			return rv.String()
		}
		return wrapped

	// any type -> unchanged
	default:
		return rv.Interface()
	}
}
