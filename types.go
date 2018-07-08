package aicra

import (
	"git.xdrm.io/go/aicra/internal/checker"
	"git.xdrm.io/go/aicra/internal/config"
	"git.xdrm.io/go/aicra/middleware"
)

// Server represents an AICRA instance featuring:
// * its type checkers
// * its middlewares
// * its configuration (controllers)
type Server struct {
	config     *config.Controller
	Params     map[string]interface{}
	Checker    *checker.TypeRegistry          // type check
	Middleware *middleware.MiddlewareRegistry // middlewares
}
