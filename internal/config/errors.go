package config

// Err allows you to create constant "const" error with type boxing.
type Err string

func (err Err) Error() string {
	return string(err)
}

const (
	// ErrRead - read error
	ErrRead = Err("cannot read config")

	// ErrUnknownMethod - unknown http method
	ErrUnknownMethod = Err("unknown HTTP method")

	// ErrFormat - invalid format
	ErrFormat = Err("invalid config format")

	// ErrPatternCollision - collision between 2 services' patterns
	ErrPatternCollision = Err("pattern collision")

	// ErrInvalidPattern - malformed service pattern
	ErrInvalidPattern = Err("malformed service path: must begin with a '/' and not end with")

	// ErrInvalidPatternBraceCapture - invalid brace capture
	ErrInvalidPatternBraceCapture = Err("invalid uri parameter")

	// ErrUnspecifiedBraceCapture - missing path brace capture
	ErrUnspecifiedBraceCapture = Err("missing uri parameter")

	// ErrUndefinedBraceCapture - missing capturing brace definition
	ErrUndefinedBraceCapture = Err("missing uri parameter definition")

	// ErrMandatoryRename - capture/query parameters must be renamed
	ErrMandatoryRename = Err("uri and query parameters must be renamed")

	// ErrMissingDescription - a service is missing its description
	ErrMissingDescription = Err("missing description")

	// ErrIllegalOptionalURIParam - uri parameter cannot optional
	ErrIllegalOptionalURIParam = Err("uri parameter cannot be optional")

	// ErrOptionalOption - cannot have optional output
	ErrOptionalOption = Err("output cannot be optional")

	// ErrMissingParamDesc - missing parameter description
	ErrMissingParamDesc = Err("missing parameter description")

	// ErrUnknownParamType - unknown parameter type
	ErrUnknownParamType = Err("unknown parameter datatype")

	// ErrIllegalParamName - illegal parameter name
	ErrIllegalParamName = Err("illegal parameter name")

	// ErrMissingParamType - missing parameter type
	ErrMissingParamType = Err("missing parameter type")

	// ErrParamNameConflict - name/rename conflict
	ErrParamNameConflict = Err("parameter name conflict")
)
