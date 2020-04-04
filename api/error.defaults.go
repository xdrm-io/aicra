package api

import "net/http"

var (
	// ErrorUnknown represents any error which cause is unknown.
	// It might also be used for debug purposes as this error
	// has to be used the less possible
	ErrorUnknown Error = -1

	// ErrorSuccess represents a generic successful service execution
	ErrorSuccess Error = 0

	// ErrorFailure is the most generic error
	ErrorFailure Error = 1

	// ErrorNoMatchFound has to be set when trying to fetch data and there is no result
	ErrorNoMatchFound Error = 2

	// ErrorAlreadyExists has to be set when trying to insert data, but identifiers or
	// unique fields already exists
	ErrorAlreadyExists Error = 3

	// ErrorCreation has to be set when there is a creation/insert error
	ErrorCreation Error = 4

	// ErrorModification has to be set when there is an update/modification error
	ErrorModification Error = 5

	// ErrorDeletion has to be set when there is a deletion/removal error
	ErrorDeletion Error = 6

	// ErrorTransaction has to be set when there is a transactional error
	ErrorTransaction Error = 7

	// ErrorUpload has to be set when a file upload failed
	ErrorUpload Error = 100

	// ErrorDownload has to be set when a file download failed
	ErrorDownload Error = 101

	// MissingDownloadHeaders has to be set when the implementation
	// of a service of type 'download' (which returns a file instead of
	// a set or output fields) is missing its HEADER field
	MissingDownloadHeaders Error = 102

	// ErrorMissingDownloadBody has to be set when the implementation
	// of a service of type 'download' (which returns a file instead of
	// a set or output fields) is missing its BODY field
	ErrorMissingDownloadBody Error = 103

	// ErrorUnknownService is set when there is no service matching
	// the http request URI.
	ErrorUnknownService Error = 200

	// ErrorUncallableService is set when there the requested service's
	// implementation (plugin file) is not found/callable
	ErrorUncallableService Error = 202

	// ErrorNotImplemented is set when a handler is not implemented yet
	ErrorNotImplemented Error = 203

	// ErrorPermission is set when there is a permission error by default
	// the api returns a permission error when the current scope (built
	// by middlewares) does not match the scope required in the config.
	// You can add your own permission policy and use this error
	ErrorPermission Error = 300

	// ErrorToken has to be set (usually in authentication middleware) to tell
	// the user that this authentication token is expired or invalid
	ErrorToken Error = 301

	// ErrorMissingParam is set when a *required* parameter is missing from the
	// http request
	ErrorMissingParam Error = 400

	// ErrorInvalidParam is set when a given parameter fails its type check as
	// defined in the config file.
	ErrorInvalidParam Error = 401

	// ErrorInvalidDefaultParam is set when an optional parameter's default value
	// does not match its type.
	ErrorInvalidDefaultParam Error = 402
)

var errorReasons = map[Error]string{
	ErrorUnknown:             "unknown error",
	ErrorSuccess:             "all right",
	ErrorFailure:             "it failed",
	ErrorNoMatchFound:        "resource not found",
	ErrorAlreadyExists:       "already exists",
	ErrorCreation:            "create error",
	ErrorModification:        "update error",
	ErrorDeletion:            "delete error",
	ErrorTransaction:         "transactional error",
	ErrorUpload:              "upload failed",
	ErrorDownload:            "download failed",
	MissingDownloadHeaders:   "download headers are missing",
	ErrorMissingDownloadBody: "download body is missing",
	ErrorUnknownService:      "unknown service",
	ErrorUncallableService:   "uncallable service",
	ErrorNotImplemented:      "not implemented",
	ErrorPermission:          "permission error",
	ErrorToken:               "token error",
	ErrorMissingParam:        "missing parameter",
	ErrorInvalidParam:        "invalid parameter",
	ErrorInvalidDefaultParam: "invalid default param",
}
var errorStatus = map[Error]int{
	ErrorUnknown:             http.StatusOK,
	ErrorSuccess:             http.StatusOK,
	ErrorFailure:             http.StatusInternalServerError,
	ErrorNoMatchFound:        http.StatusOK,
	ErrorAlreadyExists:       http.StatusOK,
	ErrorCreation:            http.StatusOK,
	ErrorModification:        http.StatusOK,
	ErrorDeletion:            http.StatusOK,
	ErrorTransaction:         http.StatusOK,
	ErrorUpload:              http.StatusInternalServerError,
	ErrorDownload:            http.StatusInternalServerError,
	MissingDownloadHeaders:   http.StatusBadRequest,
	ErrorMissingDownloadBody: http.StatusBadRequest,
	ErrorUnknownService:      http.StatusServiceUnavailable,
	ErrorUncallableService:   http.StatusServiceUnavailable,
	ErrorNotImplemented:      http.StatusNotImplemented,
	ErrorPermission:          http.StatusUnauthorized,
	ErrorToken:               http.StatusForbidden,
	ErrorMissingParam:        http.StatusBadRequest,
	ErrorInvalidParam:        http.StatusBadRequest,
	ErrorInvalidDefaultParam: http.StatusBadRequest,
}
