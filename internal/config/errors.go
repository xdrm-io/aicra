package config

// cerr allows you to create constant "const" error with type boxing.
type cerr string

func (err cerr) Error() string {
	return string(err)
}

// errRead - a problem ocurred when trying to read the configuration file
const errRead = cerr("cannot read config")

// errUnknownMethod - invalid http method
const errUnknownMethod = cerr("unknown HTTP method")

// errFormat - a invalid format has been detected
const errFormat = cerr("invalid config format")

// errPatternCollision - there is a collision between 2 services' patterns (same method)
const errPatternCollision = cerr("pattern collision")

// errInvalidPattern - a service pattern is malformed
const errInvalidPattern = cerr("must begin with a '/' and not end with")

// errInvalidPatternBraceCapture - a service pattern brace capture is invalid
const errInvalidPatternBraceCapture = cerr("invalid uri capturing braces")

// errUnspecifiedBraceCapture - a parameter brace capture is not specified in the pattern
const errUnspecifiedBraceCapture = cerr("capturing brace missing in the path")

// errMandatoryRename - capture/query parameters must have a rename
const errMandatoryRename = cerr("capture and query parameters must have a 'name'")

// errUndefinedBraceCapture - a parameter brace capture in the pattern is not defined in parameters
const errUndefinedBraceCapture = cerr("capturing brace missing input definition")

// errMissingDescription - a service is missing its description
const errMissingDescription = cerr("missing description")

// errIllegalOptionalURIParam - an URI parameter cannot be optional
const errIllegalOptionalURIParam = cerr("URI parameter cannot be optional")

// errOptionalOption - an output is optional
const errOptionalOption = cerr("output cannot be optional")

// errMissingParamDesc - a parameter is missing its description
const errMissingParamDesc = cerr("missing parameter description")

// errUnknownDataType - a parameter has an unknown datatype name
const errUnknownDataType = cerr("unknown data type")

// errIllegalParamName - a parameter has an illegal name
const errIllegalParamName = cerr("illegal parameter name")

// errMissingParamType - a parameter has an illegal type
const errMissingParamType = cerr("missing parameter type")

// errParamNameConflict - a parameter has a conflict with its name/rename field
const errParamNameConflict = cerr("name conflict for parameter")
