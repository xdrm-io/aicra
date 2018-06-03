package implement

import (
	"git.xdrm.io/xdrm-brackets/gfw/err"
	"sync"
)

type Arguments map[string]interface{}
type Controller func(Arguments, *Response) Response

type Response struct {
	data map[string]interface{}
	m    sync.Mutex
	Err  err.Error
}
