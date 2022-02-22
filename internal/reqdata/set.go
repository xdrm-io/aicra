package reqdata

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"reflect"

	"github.com/xdrm-io/aicra/internal/config"

	"mime/multipart"
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

	var (
		contentType            = req.Header.Get("Content-Type")
		mediaType, params, err = mime.ParseMediaType(contentType)
	)
	if err != nil {
		return err
	}

	switch {
	case strings.HasPrefix(mediaType, "application/json"):
		err := i.parseJSON(req)
		if err != nil {
			return err
		}

	case strings.HasPrefix(mediaType, "application/x-www-form-urlencoded"):
		err := i.parseUrlencoded(req)
		if err != nil {
			return err
		}

	case strings.HasPrefix(mediaType, "multipart/form-data"):
		err := i.parseMultipart(req, params["boundary"])
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
func (i *T) parseMultipart(req http.Request, boundary string) error {
	type Part struct {
		contentType string
		data        []byte
	}

	var (
		parts = map[string]Part{}
		mr    = multipart.NewReader(req.Body, boundary)
	)
	var firstPart = true
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		// first part is empty -> consider as empty multipart
		if firstPart && err != nil && strings.HasSuffix(err.Error(), ": EOF") {
			return nil
		}
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidMultipart, err)
		}
		firstPart = false

		data, err := ioutil.ReadAll(p)
		if err != nil {
			return fmt.Errorf("%w: %s: %s", ErrInvalidMultipart, p.FormName(), err)
		}
		parts[p.FormName()] = Part{
			contentType: p.Header.Get("Content-Type"),
			data:        data,
		}
	}

	for name, param := range i.service.Form {
		part, exist := parts[name]
		if !exist {
			continue
		}

		var isFile = len(part.contentType) > 0

		var parsed interface{}
		if isFile {
			parsed = part.data
		} else {
			parsed = parseParameter(string(part.data))
		}

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
// - other    : bypass
func parseParameter(data interface{}) interface{} {
	switch cast := data.(type) {

	// []string -> recursive
	case []string:
		if len(cast) == 0 {
			return cast
		}
		slice := make([]interface{}, len(cast))
		for i, v := range cast {
			slice[i] = parseParameter(v)
		}
		return slice

	// string -> parse as json
	case string:
		var (
			receiver map[string]interface{}
			wrapper  = fmt.Sprintf("{\"wrapped\":%s}", cast)
			err      = json.Unmarshal([]byte(wrapper), &receiver)
		)
		if err != nil {
			return cast
		}

		wrapped, ok := receiver["wrapped"]
		if !ok {
			return cast
		}
		return wrapped

	// other -> bypass
	default:
		return cast
	}
}
