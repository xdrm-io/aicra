package api

var (
	// ErrorSuccess represents a generic successful controller execution
	ErrorSuccess = func() Error { return Error{0, "all right", nil} }

	// ErrorFailure is the most generic error
	ErrorFailure = func() Error { return Error{1, "it failed", nil} }

	// ErrorUnknown represents any error which cause is unknown.
	// It might also be used for debug purposes as this error
	// has to be used the less possible
	ErrorUnknown = func() Error { return Error{-1, "", nil} }

	// ErrorNoMatchFound has to be set when trying to fetch data and there is no result
	ErrorNoMatchFound = func() Error { return Error{2, "no resource found", nil} }

	// ErrorAlreadyExists has to be set when trying to insert data, but identifiers or
	// unique fields already exists
	ErrorAlreadyExists = func() Error { return Error{3, "resource already exists", nil} }

	// ErrorConfig has to be set when there is a configuration error
	ErrorConfig = func() Error { return Error{4, "configuration error", nil} }

	// ErrorUpload has to be set when a file upload failed
	ErrorUpload = func() Error { return Error{100, "upload failed", nil} }

	// ErrorDownload has to be set when a file download failed
	ErrorDownload = func() Error { return Error{101, "download failed", nil} }

	// MissingDownloadHeaders has to be set when the implementation
	// of a controller of type 'download' (which returns a file instead of
	// a set or output fields) is missing its HEADER field
	MissingDownloadHeaders = func() Error { return Error{102, "download headers are missing", nil} }

	// ErrorMissingDownloadBody has to be set when the implementation
	// of a controller of type 'download' (which returns a file instead of
	// a set or output fields) is missing its BODY field
	ErrorMissingDownloadBody = func() Error { return Error{103, "download body is missing", nil} }

	// ErrorUnknownService is set when there is no controller matching
	// the http request URI.
	ErrorUnknownService = func() Error { return Error{200, "unknown service", nil} }

	// ErrorUnknownMethod is set when there is no method matching the
	// request's http method
	ErrorUnknownMethod = func() Error { return Error{201, "unknown method", nil} }

	// ErrorUncallableService is set when there the requested controller's
	// implementation (plugin file) is not found/callable
	// ErrorUncallableService = func() Error { return Error{202, "uncallable service", nil} }

	// ErrorUncallableMethod is set when there the requested controller's
	// implementation does not features the requested method
	// ErrorUncallableMethod = func() Error { return Error{203, "uncallable method", nil} }

	// ErrorPermission is set when there is a permission error by default
	// the api returns a permission error when the current scope (built
	// by middlewares) does not match the scope required in the config.
	// You can add your own permission policy and use this error
	ErrorPermission = func() Error { return Error{300, "permission error", nil} }

	// ErrorToken has to be set (usually in authentication middleware) to tell
	// the user that this authentication token is expired or invalid
	ErrorToken = func() Error { return Error{301, "token error", nil} }

	// ErrorMissingParam is set when a *required* parameter is missing from the
	// http request
	ErrorMissingParam = func() Error { return Error{400, "missing parameter", nil} }

	// ErrorInvalidParam is set when a given parameter fails its type check as
	// defined in the config file.
	ErrorInvalidParam = func() Error { return Error{401, "invalid parameter", nil} }

	// ErrorInvalidDefaultParam is set when an optional parameter's default value
	// does not match its type.
	ErrorInvalidDefaultParam = func() Error { return Error{402, "invalid default param", nil} }
)
