package api

import (
	"fmt"
)

// Err represents an http response error following the api format.
// These are used by the services to set the *execution status*
// directly into the response as JSON alongside response output fields.
type Err struct {
	// error code (unique)
	Code int `json:"code"`
	// error small description
	Reason string `json:"reason"`
	// associated HTTP status
	Status int
}

func (e Err) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Reason)
}
