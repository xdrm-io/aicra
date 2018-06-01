package implement

import (
	"git.xdrm.io/xdrm-brackets/gfw/err"
)

func NewResponse() *Response {
	return &Response{
		data: make(map[string]interface{}),
		Err:  err.Success,
	}
}

func (i *Response) Set(name string, value interface{}) {
	i.m.Lock()
	defer i.m.Unlock()

	i.data[name] = value
}

func (i *Response) Get(name string) interface{} {
	i.m.Lock()
	value, _ := i.data[name]
	i.m.Unlock()

	return value
}

func (i *Response) Dump() map[string]interface{} {
	i.m.Lock()
	defer i.m.Unlock()

	return i.data
}
