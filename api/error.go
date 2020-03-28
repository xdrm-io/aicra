package api

import (
	"encoding/json"
	"fmt"
)

// Error represents an http response error following the api format.
// These are used by the services to set the *execution status*
// directly into the response as JSON alongside response output fields.
type Error int

// Error implements the error interface
func (e Error) Error() string {
	// use unknown error if no reason
	reason, ok := errorReasons[e]
	if !ok {
		return ErrorUnknown.Error()
	}

	return fmt.Sprintf("[%d] %s", e, reason)
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
