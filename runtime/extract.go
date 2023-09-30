package runtime

import (
	"net/http"
	"net/url"
	"reflect"

	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/aicra/validator"
)

// ExtractURI extracts an URI parameter from an http request
func ExtractURI[T any](r *http.Request, i int, extractor validator.ExtractFunc[T]) (T, error) {
	var zero T

	fragments := config.URIFragments(r.RequestURI)
	if i >= len(fragments) {
		return zero, ErrMissingURIParameter
	}

	v, ok := extractor(fragments[i])
	if !ok {
		return zero, ErrInvalidType
	}
	return v, nil
}

// ExtractQuery extracts an Query parameter from an http request
func ExtractQuery[T any](r *http.Request, name string, extractor validator.ExtractFunc[T]) (T, error) {
	var zero T

	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return zero, err
	}

	values, ok := query[name]
	if !ok {
		return zero, ErrMissingParam
	}

	value, err := extractFromStringList[T](values)
	if err != nil {
		return zero, err
	}
	v, ok := extractor(value)
	if !ok {
		return zero, ErrInvalidType
	}
	return v, nil
}

// ExtractForm extracts a Form parameter from an http request
func ExtractForm[T any](form Form, name string, extractor validator.ExtractFunc[T]) (T, error) {
	var zero T

	switch form.typ {
	case JSON:
		value, ok := form.values[name]
		if !ok {
			return zero, ErrMissingParam
		}
		v, ok := extractor(value)
		if !ok {
			return zero, ErrInvalidType
		}
		return v, nil

	case URLEncoded:
		raw, ok := form.values[name]
		if !ok {
			return zero, ErrMissingParam
		}
		values, ok := raw.([]string)
		if !ok {
			return zero, ErrParseParameter
		}
		value, err := extractFromStringList[T](values)
		if err != nil {
			return zero, err
		}
		v, ok := extractor(value)
		if !ok {
			return zero, ErrInvalidType
		}
		return v, nil

	case Multipart:
		value, ok := form.values[name]
		if !ok {
			return zero, ErrMissingParam
		}
		v, ok := extractor(value)
		if !ok {
			return zero, ErrInvalidType
		}
		return v, nil

	}
	return zero, ErrUnhandledContentType
}

// extractFromStringList extracts values from a string list. If the expected
// type is a slice it reads all values ; otherwise it only expects a single value.
func extractFromStringList[T any](values []string) (any, error) {
	var zero T

	// expect slice
	if reflect.TypeOf(zero).Kind() == reflect.Slice {
		return values, nil
	}
	if len(values) != 1 {
		return zero, ErrParseParameter
	}
	return values[0], nil
}