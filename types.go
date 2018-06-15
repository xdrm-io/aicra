package gfw

import (
	"git.xdrm.io/go/xb-api/checker"
	"git.xdrm.io/go/xb-api/config"
)

type Server struct {
	config  *config.Controller
	Params  map[string]interface{}
	Checker *checker.TypeRegistry // type check
}
