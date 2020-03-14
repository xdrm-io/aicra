package config

// Error allows you to create constant "const" error with type boxing.
type Error string

// Error implements the error builtin interface.
func (err Error) Error() string {
	return string(err)
}

// ErrRead - a problem ocurred when trying to read the configuration file
const ErrRead = Error("cannot read config")

// ErrUnknownMethod - invalid http method
const ErrUnknownMethod = Error("unknown HTTP method")

// ErrFormat - a invalid format has been detected
const ErrFormat = Error("invalid config format")

// ErrPatternCollision - there is a collision between 2 services' patterns (same method)
const ErrPatternCollision = Error("invalid config format")

// ErrInvalidPattern - a service pattern is malformed
const ErrInvalidPattern = Error("must begin with a '/' and not end with")

// ErrInvalidPatternBracePosition - a service pattern opening/closing brace is not directly between '/'
const ErrInvalidPatternBracePosition = Error("capturing braces must be alone between slashes")

// ErrInvalidPatternOpeningBrace - a service pattern opening brace is invalid
const ErrInvalidPatternOpeningBrace = Error("opening brace already open")

// ErrInvalidPatternClosingBrace - a service pattern closing brace is invalid
const ErrInvalidPatternClosingBrace = Error("closing brace already closed")

// ErrMissingDescription - a service is missing its description
const ErrMissingDescription = Error("missing description")

// ErrMissingParamDesc - a parameter is missing its description
const ErrMissingParamDesc = Error("missing parameter description")

// ErrIllegalParamName - a parameter has an illegal name
const ErrIllegalParamName = Error("parameter name must not begin/end with '_'")

// ErrMissingParamType - a parameter has an illegal type
const ErrMissingParamType = Error("missing parameter type")

// ErrParamNameConflict - a parameter has a conflict with its name/rename field
const ErrParamNameConflict = Error("name conflict for parameter")
