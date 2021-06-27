package api

import (
	"net/http"
	"strconv"
	"strings"
)

// Err defines api errors
type Err string

// Error implements the error interface
func (e Err) Error() string {
	return strings.SplitN(string(e), ":", 2)[1]
}

// Status returns the associated http status code
func (e Err) Status() int {
	str := strings.SplitN(string(e), ":", 2)[0]
	status, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return http.StatusInternalServerError
	}
	return int(status)
}

const (
	// ErrFailure is the most generic error
	ErrFailure = Err("500:it failed")

	// ErrNotFound is thrown when trying to fetch data and there is no result
	ErrNotFound = Err("404:not found")

	// ErrAlreadyExists is thrown when trying to insert data, but identifiers or
	// unique fields already exists
	ErrAlreadyExists = Err("500:already exists")

	// ErrCreate is thrown when there is a creation/insert error
	ErrCreate = Err("500:create error")

	// ErrUpdate is thrown when there is an update/modification error
	ErrUpdate = Err("500:update error")

	// ErrDelete is thrown when there is a deletion/removal error
	ErrDelete = Err("500:delete error")

	// ErrTransaction is thrown when there is a transactional error
	ErrTransaction = Err("500:transactional error")

	// ErrUnknownService is thrown when there is no service matching
	// the http request method and URI
	ErrUnknownService = Err("503:unknown service")

	// ErrUncallableService is thrown when there the requested service's handler
	// is not found/callable
	ErrUncallableService = Err("503:uncallable service")

	// ErrNotImplemented is thrown when a handler is not implemented yet
	ErrNotImplemented = Err("501:not implemented")

	// ErrUnauthorized is thrown (usually in authentication middleware) to tell
	// the user that an authentication is required
	ErrUnauthorized = Err("401:unauthorized")

	// ErrForbidden is thrown when there is a permission error by default
	// the api returns a permission error when the current scope (built
	// by middlewares) does not match the scope required in the config.
	// You can add your own permission policy and use this error
	ErrForbidden = Err("403:forbidden")

	// ErrMissingParam is thrown when a *required* parameter is missing from the
	// http request
	ErrMissingParam = Err("400:missing parameter")

	// ErrInvalidParam is thrown when a given parameter fails its type check as
	// defined in the config file.
	ErrInvalidParam = Err("400:invalid parameter")
)

// GetErrorStatus returns the http status associated with a given error if the
// error type implements the interface{ Status() int }
func GetErrorStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	withStatus, ok := err.(interface{ Status() int })
	if !ok {
		return http.StatusInternalServerError
	}
	return withStatus.Status()
}
