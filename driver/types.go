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

// Generic tells the aicra instance to use the generic driver to load controller/middleware executables
//
// It will call an executable with the json input into the standard input (argument 1)
//    the HTTP method is send as the key _HTTP_METHOD_ (in upper case)
// The standard output must be a json corresponding to the data
//
// CONTROLLER FILE STRUCTURE
// --------------
// - the root (/) controller executable must be named  <WORKDIR>/controller/ROOT
// - the a/b/c controller executable must be named <WORKDIR>/controller/a/b/c
type Generic struct{}

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

// Import tells the aicra instance to use the import driver to load controller/middleware executables
//
// It will compile imported with the following interface :
//
// type Controller interface {
// 	Get(d response.Arguments) response.Response
// 	Post(d response.Arguments) response.Response
// 	Put(d response.Arguments) response.Response
// 	Delete(d response.Arguments) response.Response
// }
//
// CONTROLLER FILE STRUCTURE
// --------------
// - the root (/) controller executable must be named  <WORKDIR>/controller/ROOT.go
// - the a/b/c controller executable must be named <WORKDIR>/controller/a/b/c.go
//
// COMPILATION
// -----------
// The controllers/middlewares are imported and instanciated inside the main function, they are
// thus compiled with the main binary file
type Import struct {
	Controllers map[string]Controller
	Middlewares map[string]Middleware
}

// Plugin tells the aicra instance to use the plugin driver to load controller/middleware executables
//
// It will load go .so plugins with the following interface :
//
// type Plugin interface {
//		Get(d i.Arguments, r *i.Response) i.Response
//		Post(d i.Arguments, r *i.Response) i.Response
//		Put(d i.Arguments, r *i.Response) i.Response
//		Delete(d i.Arguments, r *i.Response) i.Response
// }
//
// CONTROLLER FILE STRUCTURE
// --------------
// - the root (/) controller executable must be named  <WORKDIR>/controller/ROOT/main.so
// - the a/b/c controller executable must be named <WORKDIR>/controller/a/b/c/main.so
//
// COMPILATION
// -----------
// The compilation is handled with the command-line tool `aicra <WORKDIR>`
type Plugin struct{}

// FastCGI tells the aicra instance to use the fastcgi driver to load controller/middleware executables
//
// Warning: PHP only
//
// It will use the fastcgi protocol with php at <host>:<port>
//
// CONTROLLER FILE STRUCTURE
// --------------
// - the root (/) controller executable must be named  <WORKDIR>/controller/ROOT.php
// - the a/b/c controller executable must be named <WORKDIR>/controller/a/b/c.php
type FastCGI struct {
	host string
	port string
}
