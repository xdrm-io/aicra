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

// ErrInvalidPatternBraceCapture - a service pattern brace capture is invalid
const ErrInvalidPatternBraceCapture = Error("invalid uri capturing braces")

// ErrUnspecifiedBraceCapture - a parameter brace capture is not specified in the pattern
const ErrUnspecifiedBraceCapture = Error("capturing brace missing in the path")

// ErrMissingDescription - a service is missing its description
const ErrMissingDescription = Error("missing description")

// ErrMissingParamDesc - a parameter is missing its description
const ErrMissingParamDesc = Error("missing parameter description")

// ErrUnknownDataType - a parameter has an unknown datatype name
const ErrUnknownDataType = Error("unknown data type")

// ErrIllegalParamName - a parameter has an illegal name
const ErrIllegalParamName = Error("illegal parameter name")

// ErrMissingParamType - a parameter has an illegal type
const ErrMissingParamType = Error("missing parameter type")

// ErrParamNameConflict - a parameter has a conflict with its name/rename field
const ErrParamNameConflict = Error("name conflict for parameter")
