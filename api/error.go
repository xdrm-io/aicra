package api

import (
	"fmt"
)

// Error represents an http response error following the api format.
// These are used by the controllers to set the *execution status*
// directly into the response as JSON alongside response output fields.
type Error struct {
	Code      int           `json:"code"`
	Reason    string        `json:"reason"`
	Arguments []interface{} `json:"arguments"`
}

// Put adds an argument to the error
// to be displayed back to API caller
func (e *Error) Put(arg interface{}) {

	/* (1) Make slice if not */
	if e.Arguments == nil {
		e.Arguments = make([]interface{}, 0)
	}

	/* (2) Append argument */
	e.Arguments = append(e.Arguments, arg)

}

// Implements 'error'
func (e Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Reason)
}
