package driver

import (
	"git.xdrm.io/go/aicra/err"
	"git.xdrm.io/go/aicra/response"
	"net/http"
)

// Driver defines the driver interface to load controller/middleware implementation or executables
type Driver interface {
	Name() string
	Path(string, string, string) string
	Source(string, string, string) string
	Build(string, string, string) string
	Compiled() bool

	RunController(_path []string, _method string) (func(response.Arguments) response.Response, err.Error)
	LoadMiddleware(_path string) (func(http.Request, *[]string), error)
}

// Controller is the interface that controller implementation must follow
// it is used by the 'Import' driver
type Controller interface {
	Get(d response.Arguments) response.Response
	Post(d response.Arguments) response.Response
	Put(d response.Arguments) response.Response
	Delete(d response.Arguments) response.Response
}

// Middleware is the interface that middleware implementation must follow
// it is used by the 'Import' driver
type Middleware interface {
	Inspect(http.Request, *[]string)
}
