package err

import (
	"encoding/json"
	"fmt"
)

// Error represents an http response error following the api format.
// These are used by the controllers to set the *execution status*
// directly into the response as JSON alongside response output fields.
type Error struct {
	Code      int
	Reason    string
	Arguments []interface{}
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

// MarshalJSON implements the 'json.Marshaler' interface and is used
// to generate the JSON representation of the error data
func (e Error) MarshalJSON() ([]byte, error) {

	var jsonArguments string

	/* (1) Marshal 'Arguments' if set */
	if e.Arguments != nil && len(e.Arguments) > 0 {
		argRepresentation, err := json.Marshal(e.Arguments)
		if err == nil {
			jsonArguments = fmt.Sprintf(",\"arguments\":%s", argRepresentation)
		}

	}

	/* (2) Render JSON manually */
	return []byte(fmt.Sprintf("{\"error\":%d,\"reason\":\"%s\"%s}", e.Code, e.Reason, jsonArguments)), nil

}
