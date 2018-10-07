package driver

import (
	"git.xdrm.io/go/aicra/api"
	"net/http"
)

// Driver defines the driver interface to load controller/middleware implementation or executables
type Driver interface {
	Name() string
	Path(string, string, string) string
	Source(string, string, string) string
	Build(string, string, string) string
	Compiled() bool

	LoadController(_path string) (Controller, error)
	LoadMiddleware(_path string) (Middleware, error)
	LoadChecker(_path string) (Checker, error)
}

// Checker is the interface that type checkers implementation must follow
type Checker interface {
	Match(string) bool
	Check(interface{}) bool
}

// Controller is the interface that controller implementation must follow
// it is used by the 'Import' driver
type Controller interface {
	Get(d api.Arguments) api.Response
	Post(d api.Arguments) api.Response
	Put(d api.Arguments) api.Response
	Delete(d api.Arguments) api.Response
}

// Middleware is the interface that middleware implementation must follow
// it is used by the 'Import' driver
type Middleware interface {
	Inspect(http.Request, *[]string)
}
