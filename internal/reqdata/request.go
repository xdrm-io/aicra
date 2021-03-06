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

// Request represents all data that can be extracted from an http request for a
// specific configuration service; it features:
// - uri data
// - get data
// - form data depending on the Content-Type http header
//   - 'application/json'                  => key-value pair is parsed as json into the map
//   - 'application/x-www-form-urlencoded' => standard parameters as QUERY parameters
//   - 'multipart/form-data'               => parse form-data format
type Request struct {
	req     *http.Request
	service *config.Service
	Data    map[string]interface{}
}

// NewRequest creates a new empty store.
func NewRequest(req *http.Request, service *config.Service) *Request {
	return &Request{
		req:     req,
		service: service,
		Data:    map[string]interface{}{},
	}
}

// ExtractURI parameters
func (r *Request) ExtractURI() error {
	if len(r.service.Captures) < 1 {
		return nil
	}
	uriparts := config.SplitURI(r.req.URL.Path)

	for _, capture := range r.service.Captures {
		// out of range
		if capture.Index > len(uriparts)-1 {
			return &Err{field: capture.Ref.Rename, err: ErrMissingURIParameter}
		}
		value := uriparts[capture.Index]

		if capture.Ref == nil {
			panic(fmt.Errorf("unknown uri part type: %q", capture.Name))
		}

		parsed := value
		cast, valid := capture.Ref.Validator(parsed)
		if !valid {
			return &Err{field: capture.Ref.Rename, err: ErrInvalidType}
		}
		r.Data[capture.Ref.Rename] = cast
	}
	return nil
}

// ExtractQuery data from the url query parameters
func (r *Request) ExtractQuery() error {
	if len(r.service.Query) < 1 {
		return nil
	}
	query := r.req.URL.Query()

	for name, param := range r.service.Query {
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
			parsed = values
		} else {
			// should expect at most 1 value
			if len(values) > 1 {
				return &Err{field: param.Rename, err: ErrInvalidType}
			}
			if len(values) > 0 {
				parsed = values[0]
			}
		}

		cast, valid := param.Validator(parsed)
		if !valid {
			return &Err{field: param.Rename, err: ErrInvalidType}
		}
		r.Data[param.Rename] = cast
	}

	return nil
}

// ExtractForm parameters according go the http Content-Type header
// - 'multipart/form-data'
// - 'x-www-form-urlencoded'
// - 'application/json'
func (r *Request) ExtractForm() error {
	if r.req.Method == http.MethodGet {
		return nil
	}

	var contentType = r.req.Header.Get("Content-Type")

	switch {
	case strings.HasPrefix(contentType, "application/json"):
		err := r.parseJSON()
		if err != nil {
			return err
		}

	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		err := r.parseUrlencoded()
		if err != nil {
			return err
		}

	case strings.HasPrefix(contentType, "multipart/form-data"):
		_, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			break
		}
		err = r.parseMultipart(params["boundary"])
		if err != nil {
			return err
		}
	}

	// fail on at least 1 mandatory form param when there is no body
	for _, param := range r.service.Form {
		_, exists := r.Data[param.Rename]
		if !exists && !param.Optional {
			return &Err{field: param.Rename, err: ErrMissingRequiredParam}
		}
	}
	return nil
}

// parseJSON parses JSON from the request body inside 'Form'
// and 'Set'
func (r *Request) parseJSON() error {
	var parsed map[string]interface{}

	decoder := json.NewDecoder(r.req.Body)
	err := decoder.Decode(&parsed)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%s: %w", err, ErrInvalidJSON)
	}

	for name, param := range r.service.Form {
		value, exist := parsed[name]

		if !exist {
			continue
		}

		cast, valid := param.Validator(value)
		if !valid {
			return &Err{field: param.Rename, err: ErrInvalidType}
		}
		r.Data[param.Rename] = cast
	}

	return nil
}

// parseUrlencoded parses urlencoded from the request body inside 'Form'
// and 'Set'
func (r *Request) parseUrlencoded() error {
	body, err := io.ReadAll(r.req.Body)
	if err != nil {
		return err
	}

	form := make(Query)
	if err := form.Parse(string(body)); err != nil {
		return err
	}
	// io.WriteString(os.Stdout, fmt.Sprintf("query(%s) -> %v\n", r.req.URL.RawQuery, form))

	for name, param := range r.service.Form {
		values, exist := form[name]

		if !exist {
			continue
		}

		var parsed interface{}

		// consider slice only if we expect a slice, otherwise, only take the first parameter
		if param.GoType.Kind() == reflect.Slice {
			parsed = values
		} else if len(values) > 0 {
			// should expect at most 1 value
			if len(values) > 1 {
				return &Err{field: param.Rename, err: ErrInvalidType}
			}
			if len(values) > 0 {
				parsed = values[0]
			}
		}

		cast, valid := param.Validator(parsed)
		if !valid {
			return &Err{field: param.Rename, err: ErrInvalidType}
		}
		r.Data[param.Rename] = cast
	}

	return nil
}

// parseMultipart parses multi-part from the request body inside 'Form'
// and 'Set'
func (r *Request) parseMultipart(boundary string) error {
	type Part struct {
		contentType string
		data        []byte
	}

	var (
		parts = map[string]Part{}
		mr    = multipart.NewReader(r.req.Body, boundary)
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

	for name, param := range r.service.Form {
		part, exist := parts[name]
		if !exist {
			continue
		}

		var isFile = len(part.contentType) > 0

		var parsed interface{}
		if isFile {
			parsed = part.data
		} else {
			parsed = string(part.data)
		}

		cast, valid := param.Validator(parsed)
		if !valid {
			return &Err{field: param.Rename, err: ErrInvalidType}
		}
		r.Data[param.Rename] = cast
	}

	return nil
}
