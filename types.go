package gfw

import (
	"git.xdrm.io/xdrm-brackets/gfw/checker"
	"git.xdrm.io/xdrm-brackets/gfw/config"
)

type Server struct {
	config  *config.Controller
	Params  map[string]interface{}
	Checker *checker.TypeRegistry // type check
}
