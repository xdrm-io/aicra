package api

import (
	"git.xdrm.io/go/aicra/err"
)

// Arguments contains all key-value arguments
type Arguments map[string]interface{}

// Response represents an API response to be sent
type Response struct {
	data map[string]interface{}
	Err  err.Error
}
