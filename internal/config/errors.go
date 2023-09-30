package config

// Err allows you to create constant "const" error with type boxing.
type Err string

func (err Err) Error() string {
	return string(err)
}

const (
	// ErrPackageMissing - package is empty
	ErrPackageMissing = Err("missing 'package'")

	// ErrValidatorsMissing - validators is empty
	ErrValidatorsMissing = Err("missing 'validators'")

	// ErrEndpointsMissing - endpoints is empty
	ErrEndpointsMissing = Err("missing 'endpoints'")

	// ErrImportAliasCharset - import alias uses invalid characters
	ErrImportAliasCharset = Err("invalid import name charset")

	// ErrImportPathCharset - import path uses invalid characters
	ErrImportPathCharset = Err("invalid import path charset")

	// ErrImportTwice - import path used twice
	ErrImportTwice = Err("import path cannot appear twice")

	// ErrImportReserved - import name already used internally
	ErrImportReserved = Err("import name is reserved")

	// ErrMethodUnknown - unknown http method
	ErrMethodUnknown = Err("unknown HTTP method")

	// ErrPatternCollision - collision between 2 endpoints' patterns
	ErrPatternCollision = Err("pattern collision")

	// ErrPatternInvalid - malformed endpoint pattern
	ErrPatternInvalid = Err("malformed endpoint path: must begin with a '/' and not end with")

	// ErrPatternInvalidBraceCapture - invalid brace capture
	ErrPatternInvalidBraceCapture = Err("invalid uri parameter")

	// ErrBraceCaptureUnspecified - missing path brace capture
	ErrBraceCaptureUnspecified = Err("missing uri parameter")

	// ErrBraceCaptureUndefined - missing capturing brace definition
	ErrBraceCaptureUndefined = Err("missing uri parameter definition")

	// ErrRenameMandatory - capture/query parameters must be renamed
	ErrRenameMandatory = Err("uri and query parameters must be renamed")

	// ErrRenameUnexported - form parameters must be renamed when unexported
	ErrRenameUnexported = Err("form parameters must be renamed when unexported (starting with lowercase)")

	// ErrNameMissing - a endpoint is missing its name
	ErrNameMissing = Err("missing name")

	// ErrNameUnexported - an endpoint name is unexported
	ErrNameUnexported = Err("name is unexported, must start with uppercase")

	// ErrNameInvalid - an endpoint name contains invalid characters
	ErrNameInvalid = Err("name contains illegal characters")

	// ErrDescMissing - a endpoint is missing its description
	ErrDescMissing = Err("missing description")

	// ErrParamOptionalIllegalURI - uri parameter cannot optional
	ErrParamOptionalIllegalURI = Err("uri parameter cannot be optional")

	// ErrOutputOptional - cannot have optional output
	ErrOutputOptional = Err("output cannot be optional")

	// ErrOutputURIForbidden - cannot have URI output
	ErrOutputURIForbidden = Err("output cannot be an uri parameter")

	// ErrOutputQueryForbidden - cannot have Query output
	ErrOutputQueryForbidden = Err("output cannot be an query parameter")

	// ErrParamTypeUnknown - unknown parameter type
	ErrParamTypeUnknown = Err("unknown parameter datatype")

	// ErrParamNameIllegal - illegal parameter name
	ErrParamNameIllegal = Err("illegal parameter name")

	// ErrParamTypeMissing - missing parameter type
	ErrParamTypeMissing = Err("missing parameter type")

	// ErrParamRenameInvalid - invalid or unexported parameter rename
	ErrParamRenameInvalid = Err("name is unexported, must start with uppercase")

	// ErrParamTypeInvalid - parameter type syntax is invalid
	ErrParamTypeInvalid = Err("invalid parameter type syntax")

	// ErrParamTypeParamsInvalid - invalid validator params
	ErrParamTypeParamsInvalid = Err("invalid parameter validator params")

	// ErrParamNameConflict - name/rename conflict
	ErrParamNameConflict = Err("parameter name conflict")
)
