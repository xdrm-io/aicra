package response

import (
	"git.xdrm.io/go/aicra/err"
)

// New creates an empty response
func New() *Response {
	return &Response{
		data: make(map[string]interface{}),
		Err:  err.Success,
	}
}

// Set adds/overrides a new response field
func (i *Response) Set(name string, value interface{}) {
	i.data[name] = value
}

// Get gets a response field
func (i *Response) Get(name string) interface{} {
	value, _ := i.data[name]

	return value
}

// Dump gets all key/value pairs
func (i *Response) Dump() map[string]interface{} {
	return i.data
}
