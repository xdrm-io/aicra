package reqdata

import "fmt"

// cerr defines const-enabled errors with type boxing
type cerr string

// Error implements error
func (err cerr) Error() string {
	return string(err)
}

const (
	// ErrUnknownType is returned when encountering an unknown type
	ErrUnknownType = cerr("unknown type")

	// ErrInvalidMultipart is returned when multipart parse failed
	ErrInvalidMultipart = cerr("invalid multipart")

	// ErrParseParameter is returned when a parameter fails when parsing
	ErrParseParameter = cerr("cannot parse parameter")

	// ErrInvalidJSON is returned when json parse failed
	ErrInvalidJSON = cerr("invalid json")

	// ErrMissingRequiredParam - required param is missing
	ErrMissingRequiredParam = cerr("missing required param")

	// ErrInvalidType - parameter value does not satisfy its type
	ErrInvalidType = cerr("invalid type")

	// ErrMissingURIParameter - missing an URI parameter
	ErrMissingURIParameter = cerr("missing URI parameter")
)

// Err defines errors for request data
type Err struct {
	field string
	err   cerr
}

// Error implements error
func (err Err) Error() string {
	return fmt.Sprintf("%s: %s", err.field, err.err)
}

// Field returns the field associated with the error
func (err Err) Field() string {
	return err.field
}

// Unwrap implements errors.Unwrap
func (err Err) Unwrap() error {
	return err.err
}
