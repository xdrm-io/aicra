package reqdata

// cerr allows you to create constant "const" error with type boxing.
type cerr string

// Error implements the error builtin interface.
func (err cerr) Error() string {
	return string(err)
}

// ErrUnknownType is returned when encountering an unknown type
const ErrUnknownType = cerr("unknown type")

// ErrInvalidMultipart is returned when multipart parse failed
const ErrInvalidMultipart = cerr("invalid multipart")

// ErrParseParameter is returned when a parameter fails when parsing
const ErrParseParameter = cerr("cannot parse parameter")

// ErrInvalidJSON is returned when json parse failed
const ErrInvalidJSON = cerr("invalid json")

// ErrMissingRequiredParam - required param is missing
const ErrMissingRequiredParam = cerr("missing required param")

// ErrInvalidType - parameter value does not satisfy its type
const ErrInvalidType = cerr("invalid type")

// ErrMissingURIParameter - missing an URI parameter
const ErrMissingURIParameter = cerr("missing URI parameter")
