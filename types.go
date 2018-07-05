package gfw

import (
	"git.xdrm.io/go/aicra/checker"
	"git.xdrm.io/go/aicra/config"
)

type Server struct {
	config  *config.Controller
	Params  map[string]interface{}
	Checker *checker.TypeRegistry // type check
}
