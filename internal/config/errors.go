package config

// cerr allows you to create constant "const" error with type boxing.
type cerr string

// Error implements the error builtin interface.
func (err cerr) Error() string {
	return string(err)
}

// ErrRead - a problem ocurred when trying to read the configuration file
const ErrRead = cerr("cannot read config")

// ErrUnknownMethod - invalid http method
const ErrUnknownMethod = cerr("unknown HTTP method")

// ErrFormat - a invalid format has been detected
const ErrFormat = cerr("invalid config format")

// ErrPatternCollision - there is a collision between 2 services' patterns (same method)
const ErrPatternCollision = cerr("pattern collision")

// ErrInvalidPattern - a service pattern is malformed
const ErrInvalidPattern = cerr("must begin with a '/' and not end with")

// ErrInvalidPatternBraceCapture - a service pattern brace capture is invalid
const ErrInvalidPatternBraceCapture = cerr("invalid uri capturing braces")

// ErrUnspecifiedBraceCapture - a parameter brace capture is not specified in the pattern
const ErrUnspecifiedBraceCapture = cerr("capturing brace missing in the path")

// ErrMandatoryRename - capture/query parameters must have a rename
const ErrMandatoryRename = cerr("capture and query parameters must have a 'name'")

// ErrUndefinedBraceCapture - a parameter brace capture in the pattern is not defined in parameters
const ErrUndefinedBraceCapture = cerr("capturing brace missing input definition")

// ErrMissingDescription - a service is missing its description
const ErrMissingDescription = cerr("missing description")

// ErrIllegalOptionalURIParam - an URI parameter cannot be optional
const ErrIllegalOptionalURIParam = cerr("URI parameter cannot be optional")

// ErrOptionalOption - an output is optional
const ErrOptionalOption = cerr("output cannot be optional")

// ErrMissingParamDesc - a parameter is missing its description
const ErrMissingParamDesc = cerr("missing parameter description")

// ErrUnknownDataType - a parameter has an unknown datatype name
const ErrUnknownDataType = cerr("unknown data type")

// ErrIllegalParamName - a parameter has an illegal name
const ErrIllegalParamName = cerr("illegal parameter name")

// ErrMissingParamType - a parameter has an illegal type
const ErrMissingParamType = cerr("missing parameter type")

// ErrParamNameConflict - a parameter has a conflict with its name/rename field
const ErrParamNameConflict = cerr("name conflict for parameter")
