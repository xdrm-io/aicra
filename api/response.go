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
	Headers http.Header
	Err     Error
}

// NewResponse creates an empty response
func NewResponse() *Response {
	return &Response{
		Data: make(ResponseData),
		Err:  ErrorFailure(),
	}
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

type jsonResponse struct {
	Error
	ResponseData
}

// MarshalJSON implements the 'json.Marshaler' interface and is used
// to generate the JSON representation of the response
func (i *Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonResponse{i.Err, i.Data})
}
