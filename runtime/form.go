package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

type contentType string

// supported content types
const (
	JSON       = contentType("application/json")
	URLEncoded = contentType("application/x-www-form-urlencoded")
	Multipart  = contentType("multipart/form-data")
)

// Form represents a form extracted from an http request.
type Form struct {
	typ    contentType
	values map[string]interface{}
}

// ParseForm parses the body to be ready for parameter extraction
//
// Content-Type header
// - 'multipart/form-data'
// - 'x-www-form-urlencoded'
// - 'application/json'
func ParseForm(r *http.Request) (*Form, error) {
	if r.Method == http.MethodGet {
		return nil, nil
	}
	ct := r.Header.Get("Content-Type")

	switch {
	case strings.HasPrefix(ct, string(JSON)):
		return parseJSON(r.Body)

	case strings.HasPrefix(ct, string(URLEncoded)):
		return parseUrlencoded(r.Body)

	case strings.HasPrefix(ct, string(Multipart)):
		_, params, err := mime.ParseMediaType(ct)
		if err != nil {
			break
		}
		return parseMultipart(r.Body, params["boundary"])
	}
	return nil, ErrUnhandledContentType
}

// parseJSON parses JSON from the request body inside 'Form'
// and 'Set'
func parseJSON(reader io.Reader) (*Form, error) {
	form := &Form{
		typ:    JSON,
		values: make(map[string]interface{}),
	}
	err := json.NewDecoder(reader).Decode(&form.values)
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", err, ErrInvalidJSON)
	}
	return form, nil
}

// parseUrlencoded parses urlencoded from the request body inside 'Form'
// and 'Set'
func parseUrlencoded(reader io.Reader) (*Form, error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	query, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}

	form := &Form{
		typ:    URLEncoded,
		values: make(map[string]interface{}),
	}
	for name, values := range query {
		form.values[name] = values
	}
	return form, nil
}

// parseMultipart parses multi-part from the request body inside 'Form'
// and 'Set'
func parseMultipart(r io.Reader, boundary string) (*Form, error) {
	var mr = multipart.NewReader(r, boundary)

	form := &Form{
		typ:    Multipart,
		values: make(map[string]interface{}),
	}
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidMultipart, err)
		}

		data, err := ioutil.ReadAll(p)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %w", ErrInvalidMultipart, p.FormName(), err)
		}
		form.values[p.FormName()] = data
	}
	return form, nil
}
