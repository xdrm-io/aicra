package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Error represents an http response error following the api format.
// These are used by the services to set the *execution status*
// directly into the response as JSON alongside response output fields.
type Error int

func (e Error) Error() string {
	reason, ok := errorReasons[e]
	if !ok {
		return ErrorUnknown.Error()
	}
	return fmt.Sprintf("[%d] %s", e, reason)
}

// Status returns the associated HTTP status code
func (e Error) Status() int {
	status, ok := errorStatus[e]
	if !ok {
		return http.StatusOK
	}
	return status
}

// MarshalJSON implements encoding/json.Marshaler interface
func (e Error) MarshalJSON() ([]byte, error) {
	// use unknown error if no reason
	reason, ok := errorReasons[e]
	if !ok {
		return ErrorUnknown.MarshalJSON()
	}

	// format to proper struct
	formatted := struct {
		Code   int    `json:"code"`
		Reason string `json:"reason"`
	}{
		Code:   int(e),
		Reason: reason,
	}

	return json.Marshal(formatted)
}
