package err

import (
	"encoding/json"
	"fmt"
)

type Error struct {
	Code      int
	Reason    string
	Arguments []interface{}
}

// BindArgument adds an argument to the error
// to be displayed back to API caller
func (e *Error) BindArgument(arg interface{}) {

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

// Implements json.Marshaler
func (e Error) MarshalJSON() ([]byte, error) {

	var json_arguments string

	/* (1) Marshal 'Arguments' if set */
	if e.Arguments != nil && len(e.Arguments) > 0 {
		arg_representation, err := json.Marshal(e.Arguments)
		if err == nil {
			json_arguments = fmt.Sprintf(",\"arguments\":%s", arg_representation)
		}

	}

	/* (2) Render JSON manually */
	return []byte(fmt.Sprintf("{\"error\":%d,\"reason\":\"%s\"%s}", e.Code, e.Reason, json_arguments)), nil

}
