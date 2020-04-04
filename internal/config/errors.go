package config

// cerr allows you to create constant "const" error with type boxing.
type cerr string

func (err cerr) Error() string {
	return string(err)
}

// errRead - read error
const errRead = cerr("cannot read config")

// errUnknownMethod - unknown http method
const errUnknownMethod = cerr("unknown HTTP method")

// errFormat - invalid format
const errFormat = cerr("invalid config format")

// errPatternCollision - collision between 2 services' patterns
const errPatternCollision = cerr("pattern collision")

// errInvalidPattern - malformed service pattern
const errInvalidPattern = cerr("malformed service path: must begin with a '/' and not end with")

// errInvalidPatternBraceCapture - invalid brace capture
const errInvalidPatternBraceCapture = cerr("invalid uri parameter")

// errUnspecifiedBraceCapture - missing path brace capture
const errUnspecifiedBraceCapture = cerr("missing uri parameter")

// errUndefinedBraceCapture - missing capturing brace definition
const errUndefinedBraceCapture = cerr("missing uri parameter definition")

// errMandatoryRename - capture/query parameters must be renamed
const errMandatoryRename = cerr("uri and query parameters must be renamed")

// errMissingDescription - a service is missing its description
const errMissingDescription = cerr("missing description")

// errIllegalOptionalURIParam - uri parameter cannot optional
const errIllegalOptionalURIParam = cerr("uri parameter cannot be optional")

// errOptionalOption - cannot have optional output
const errOptionalOption = cerr("output cannot be optional")

// errMissingParamDesc - missing parameter description
const errMissingParamDesc = cerr("missing parameter description")

// errUnknownDataType - unknown parameter datatype
const errUnknownDataType = cerr("unknown parameter datatype")

// errIllegalParamName - illegal parameter name
const errIllegalParamName = cerr("illegal parameter name")

// errMissingParamType - missing parameter type
const errMissingParamType = cerr("missing parameter type")

// errParamNameConflict - name/rename conflict
const errParamNameConflict = cerr("parameter name conflict")
