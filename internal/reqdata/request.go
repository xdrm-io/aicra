package reqdata

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"reflect"
	"sync"

	"github.com/xdrm-io/aicra/internal/config"

	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

var mapPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{}, 8)
	},
}

// Request represents all data that can be extracted from an http request for a
// specific configuration service; it features:
// - uri data
// - get data
// - form data depending on the Content-Type http header
//   - 'application/json'                  => key-value pair is parsed as json into the map
//   - 'application/x-www-form-urlencoded' => standard parameters as QUERY parameters
//   - 'multipart/form-data'               => parse form-data format
type Request struct {
	service *config.Service
	Data    map[string]interface{}
}

// NewRequest creates a new empty store.
func NewRequest(service *config.Service) *Request {
	r := &Request{
		service: service,
		Data:    mapPool.Get().(map[string]interface{}),
	}
	// clear previous map
	for k := range r.Data {
		delete(r.Data, k)
	}
	return r
}

// Release the request ; no method or attribute shall be used after this call on
// the same request
func (r *Request) Release() {
	mapPool.Put(r.Data)
}

// ExtractURI parameters
func (r *Request) ExtractURI(req *http.Request) error {
	if len(r.service.Captures) < 1 {
		return nil
	}
	uriparts := config.SplitURI(req.URL.Path)

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
func (r *Request) ExtractQuery(req *http.Request) error {
	if len(r.service.Query) < 1 {
		return nil
	}
	query, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return err
	}

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
func (r *Request) ExtractForm(req *http.Request) error {
	if req.Method == http.MethodGet {
		return nil
	}

	var contentType = req.Header.Get("Content-Type")

	switch {
	case strings.HasPrefix(contentType, "application/json"):
		err := r.parseJSON(req.Body)
		if err != nil {
			return err
		}

	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		err := r.parseUrlencoded(req.Body)
		if err != nil {
			return err
		}

	case strings.HasPrefix(contentType, "multipart/form-data"):
		_, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			break
		}
		err = r.parseMultipart(req.Body, params["boundary"])
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
func (r *Request) parseJSON(reader io.Reader) error {
	var parsed map[string]interface{}

	decoder := json.NewDecoder(reader)
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
func (r *Request) parseUrlencoded(reader io.Reader) error {
	body, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	query, err := url.ParseQuery(string(body))
	if err != nil {
		return err
	}

	for name, param := range r.service.Form {
		values, exist := query[name]

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
func (r *Request) parseMultipart(reader io.Reader, boundary string) error {
	type Part struct {
		contentType string
		data        []byte
	}

	var (
		parts = make(map[string]Part, len(r.service.Form))
		mr    = multipart.NewReader(reader, boundary)
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
