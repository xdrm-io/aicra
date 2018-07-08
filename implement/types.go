package implement

import (
	"git.xdrm.io/go/aicra/err"
)

type Arguments map[string]interface{}
type Controller func(Arguments, *Response) Response

type Response struct {
	data map[string]interface{}
	Err  err.Error
}
