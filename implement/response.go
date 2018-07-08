package implement

import (
	"git.xdrm.io/go/aicra/err"
)

func NewResponse() *Response {
	return &Response{
		data: make(map[string]interface{}),
		Err:  err.Success,
	}
}

func (i *Response) Set(name string, value interface{}) {
	i.data[name] = value
}

func (i *Response) Get(name string) interface{} {
	value, _ := i.data[name]

	return value
}

func (i *Response) Dump() map[string]interface{} {
	return i.data
}
