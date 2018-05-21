package gfw

import (
	"git.xdrm.io/gfw/internal/config"
)

type Server struct {
	config *config.Controller
	Params map[string]interface{}
	err    Err
}

type Request struct {
	Uri      []string
	FormData map[string]interface{}
	GetData  map[string]interface{}
}
