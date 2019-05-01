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

// SetData adds/overrides a new response field
func (i *Response) SetData(name string, value interface{}) {
	i.Data[name] = value
}

// GetData gets a response field
func (i *Response) GetData(name string) interface{} {
	value, _ := i.Data[name]

	return value
}

// MarshalJSON implements the 'json.Marshaler' interface and is used
// to generate the JSON representation of the response
func (i *Response) MarshalJSON() ([]byte, error) {
	fmt := make(map[string]interface{})

	for k, v := range i.Data {
		fmt[k] = v
	}

	fmt["error"] = i.Err

	return json.Marshal(fmt)
}

// Write writes to an HTTP response.
func (i *Response) Write(w http.ResponseWriter) error {
	w.WriteHeader(i.Status)
	w.Header().Add("Content-Type", "application/json")

	fmt, err := json.Marshal(i)
	if err != nil {
		return err
	}
	w.Write(fmt)

	return nil
}
