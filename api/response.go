package api

import (
	"encoding/json"
	"net/http"
)

// ResponseData defines format for response parameters to return
type ResponseData map[string]interface{}

// Response represents an API response to be sent
type Response struct {
	Data    ResponseData
	Status  int
	Headers http.Header
	Err     Error
}

// NewResponse creates an empty response. An optional error can be passed as its first argument.
func NewResponse(errors ...Error) *Response {
	res := &Response{
		Status:  http.StatusOK,
		Data:    make(ResponseData),
		Err:     ErrorFailure(),
		Headers: make(http.Header),
	}

	// optional error
	if len(errors) == 1 {
		res.Err = errors[0]
	}

	return res
}

// WrapError sets the error from a base error with error arguments.
func (res *Response) WrapError(baseError Error, arguments ...interface{}) {
	if len(arguments) > 0 {
		baseError.SetArguments(arguments[0], arguments[1:])
	}
	res.Err = baseError
}

// SetData adds/overrides a new response field
func (res *Response) SetData(name string, value interface{}) {
	res.Data[name] = value
}

// GetData gets a response field
func (res *Response) GetData(name string) interface{} {
	value, _ := res.Data[name]

	return value
}

// MarshalJSON implements the 'json.Marshaler' interface and is used
// to generate the JSON representation of the response
func (res *Response) MarshalJSON() ([]byte, error) {
	fmt := make(map[string]interface{})

	for k, v := range res.Data {
		fmt[k] = v
	}

	fmt["error"] = res.Err

	return json.Marshal(fmt)
}

// Write writes to an HTTP response.
func (res *Response) Write(w http.ResponseWriter) error {
	w.WriteHeader(res.Status)
	w.Header().Add("Content-Type", "application/json")

	fmt, err := json.Marshal(res)
	if err != nil {
		return err
	}
	w.Write(fmt)

	return nil
}
