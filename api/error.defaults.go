package api

import "net/http"

var (
	// ErrUnknown represents any error which cause is unknown.
	// It might also be used for debug purposes as this error
	// has to be used the less possible
	ErrUnknown = Err{-1, "unknown error", http.StatusOK}

	// ErrSuccess represents a generic successful service execution
	ErrSuccess = Err{0, "all right", http.StatusOK}

	// ErrFailure is the most generic error
	ErrFailure = Err{1, "it failed", http.StatusInternalServerError}

	// ErrNoMatchFound is set when trying to fetch data and there is no result
	ErrNoMatchFound = Err{2, "resource not found", http.StatusOK}

	// ErrAlreadyExists is set when trying to insert data, but identifiers or
	// unique fields already exists
	ErrAlreadyExists = Err{3, "already exists", http.StatusOK}

	// ErrCreation is set when there is a creation/insert error
	ErrCreation = Err{4, "create error", http.StatusOK}

	// ErrModification is set when there is an update/modification error
	ErrModification = Err{5, "update error", http.StatusOK}

	// ErrDeletion is set when there is a deletion/removal error
	ErrDeletion = Err{6, "delete error", http.StatusOK}

	// ErrTransaction is set when there is a transactional error
	ErrTransaction = Err{7, "transactional error", http.StatusOK}

	// ErrUpload is set when a file upload failed
	ErrUpload = Err{100, "upload failed", http.StatusInternalServerError}

	// ErrDownload is set when a file download failed
	ErrDownload = Err{101, "download failed", http.StatusInternalServerError}

	// MissingDownloadHeaders is set when the implementation
	// of a service of type 'download' (which returns a file instead of
	// a set or output fields) is missing its HEADER field
	MissingDownloadHeaders = Err{102, "download headers are missing", http.StatusBadRequest}

	// ErrMissingDownloadBody is set when the implementation
	// of a service of type 'download' (which returns a file instead of
	// a set or output fields) is missing its BODY field
	ErrMissingDownloadBody = Err{103, "download body is missing", http.StatusBadRequest}

	// ErrUnknownService is set when there is no service matching
	// the http request URI.
	ErrUnknownService = Err{200, "unknown service", http.StatusServiceUnavailable}

	// ErrUncallableService is set when there the requested service's
	// implementation (plugin file) is not found/callable
	ErrUncallableService = Err{202, "uncallable service", http.StatusServiceUnavailable}

	// ErrNotImplemented is set when a handler is not implemented yet
	ErrNotImplemented = Err{203, "not implemented", http.StatusNotImplemented}

	// ErrPermission is set when there is a permission error by default
	// the api returns a permission error when the current scope (built
	// by middlewares) does not match the scope required in the config.
	// You can add your own permission policy and use this error
	ErrPermission = Err{300, "permission error", http.StatusUnauthorized}

	// ErrToken is set (usually in authentication middleware) to tell
	// the user that this authentication token is expired or invalid
	ErrToken = Err{301, "token error", http.StatusForbidden}

	// ErrMissingParam is set when a *required* parameter is missing from the
	// http request
	ErrMissingParam = Err{400, "missing parameter", http.StatusBadRequest}

	// ErrInvalidParam is set when a given parameter fails its type check as
	// defined in the config file.
	ErrInvalidParam = Err{401, "invalid parameter", http.StatusBadRequest}

	// ErrInvalidDefaultParam is set when an optional parameter's default value
	// does not match its type.
	ErrInvalidDefaultParam = Err{402, "invalid default param", http.StatusBadRequest}
)
