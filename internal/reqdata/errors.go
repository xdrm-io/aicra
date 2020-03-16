package reqdata

// Error allows you to create constant "const" error with type boxing.
type Error string

// Error implements the error builtin interface.
func (err Error) Error() string {
	return string(err)
}

// ErrUnknownType is returned when encountering an unknown type
const ErrUnknownType = Error("unknown type")

// ErrInvalidJSON is returned when json parse failed
const ErrInvalidJSON = Error("invalid json")

// ErrInvalidRootType is returned when json is a map
const ErrInvalidRootType = Error("invalid json root type")

// ErrInvalidParamName - parameter has an invalid
const ErrInvalidParamName = Error("invalid parameter name")

// ErrMissingRequiredParam - required param is missing
const ErrMissingRequiredParam = Error("missing required param")

// ErrInvalidType - parameter value does not satisfy its type
const ErrInvalidType = Error("invalid type")

// ErrMissingURIParameter - missing an URI parameter
const ErrMissingURIParameter = Error("missing URI parameter")
