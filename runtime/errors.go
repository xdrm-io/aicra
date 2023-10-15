package runtime

// Err defines const-enabled errors with type boxing
type Err string

// Error implements error
func (err Err) Error() string {
	return string(err)
}

const (
	// ErrInvalidMultipart is returned when multipart parse failed
	ErrInvalidMultipart = Err("invalid multipart")

	// ErrParseParameter is returned when a parameter fails when parsing
	ErrParseParameter = Err("cannot parse parameter")

	// ErrInvalidJSON is returned when json parse failed
	ErrInvalidJSON = Err("invalid json")
	// ErrInvalidURLEncoded is returned when urlencoded parse failed
	ErrInvalidURLEncoded = Err("invalid url encoded")

	// ErrMissingParam - param is missing
	ErrMissingParam = Err("missing param")

	// ErrInvalidType - parameter value does not satisfy its type
	ErrInvalidType = Err("invalid type")

	// ErrMissingURIParameter - missing an URI parameter
	ErrMissingURIParameter = Err("missing URI parameter")

	// ErrUnhandledContentType - when an unknown Content-Type is encountered
	ErrUnhandledContentType = Err("unhandled Content-Type header")
)
