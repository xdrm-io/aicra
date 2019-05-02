package api

import (
	"fmt"
)

// Error represents an http response error following the api format.
// These are used by the services to set the *execution status*
// directly into the response as JSON alongside response output fields.
type Error struct {
	Code      int           `json:"code"`
	Reason    string        `json:"reason"`
	Arguments []interface{} `json:"arguments"`
}

// SetArguments set one or multiple arguments to the error
// to be displayed back to API caller
func (e *Error) SetArguments(arg0 interface{}, args ...interface{}) {

	// 1. clear arguments */
	e.Arguments = make([]interface{}, 0)

	// 2. add arg[0]
	e.Arguments = append(e.Arguments, arg0)

	// 3. add optional other arguments
	for _, arg := range args {
		e.Arguments = append(e.Arguments, arg)
	}

}

// Implements 'error'
func (e Error) Error() string {
	if e.Arguments == nil || len(e.Arguments) < 1 {
		return fmt.Sprintf("[%d] %s", e.Code, e.Reason)
	}

	return fmt.Sprintf("[%d] %s (%v)", e.Code, e.Reason, e.Arguments)
}
