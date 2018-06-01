package gfw

import (
	"git.xdrm.io/xdrm-brackets/gfw/checker"
	"git.xdrm.io/xdrm-brackets/gfw/config"
	"git.xdrm.io/xdrm-brackets/gfw/err"
)

type Server struct {
	config  *config.Controller
	Params  map[string]interface{}
	Checker *checker.TypeRegistry // type check
	err     err.Error
}
