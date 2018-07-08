package err

var (
	// Success represents a generic successful controller execution
	Success = Error{0, "all right", nil}

	// Failure is the more generic error
	Failure = Error{1, "it failed", nil}

	// Unknown represents any error which cause is unknown.
	// It might also be used for debug purposes as this error
	// has to be used the less possible
	Unknown = Error{-1, "", nil}

	// NoMatchFound has to be set when trying to fetch data and there is no result
	NoMatchFound = Error{2, "no resource found", nil}

	// AlreadyExists has to be set when trying to insert data, but identifiers or
	// unique fields already exists
	AlreadyExists = Error{3, "resource already exists", nil}

	// Config has to be set when there is a configuration error
	Config = Error{4, "configuration error", nil}

	// Upload has to be set when a file upload failed
	Upload = Error{100, "upload failed", nil}

	// Download has to be set when a file download failed
	Download = Error{101, "download failed", nil}

	// MissingDownloadHeaders has to be set when the implementation
	// of a controller of type 'download' (which returns a file instead of
	// a set or output fields) is missing its HEADER field
	MissingDownloadHeaders = Error{102, "download headers are missing", nil}

	// MissingDownloadBody has to be set when the implementation
	// of a controller of type 'download' (which returns a file instead of
	// a set or output fields) is missing its BODY field
	MissingDownloadBody = Error{103, "download body is missing", nil}

	// UnknownController is set when there is no controller matching
	// the http request URI.
	UnknownController = Error{200, "unknown controller", nil}

	// UnknownMethod is set when there is no method matching the
	// request's http method
	UnknownMethod = Error{201, "unknown method", nil}

	// UncallableController is set when there the requested controller's
	// implementation (plugin file) is not found/callable
	UncallableController = Error{202, "uncallable controller", nil}

	// UncallableMethod is set when there the requested controller's
	// implementation does not features the requested method
	UncallableMethod = Error{203, "uncallable method", nil}

	// Permission is set when there is a permission error by default
	// the api returns a permission error when the current scope (built
	// by middlewares) does not match the scope required in the config.
	// You can add your own permission policy and use this error
	Permission = Error{300, "permission error", nil}

	// Token has to be set (usually in authentication middleware) to tell
	// the user that this authentication token is expired or invalid
	Token = Error{301, "token error", nil}

	// MissingParam is set when a *required* parameter is missing from the
	// http request
	MissingParam = Error{400, "missing parameter", nil}

	// InvalidParam is set when a given parameter fails its type check as
	// defined in the config file.
	InvalidParam = Error{401, "invalid parameter", nil}

	// InvalidDefaultParam is set when an optional parameter's default value
	// does not match its type.
	InvalidDefaultParam = Error{402, "invalid default param", nil}
)
