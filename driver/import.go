package driver

import (
	"errors"
	"fmt"
	e "git.xdrm.io/go/aicra/err"
	"git.xdrm.io/go/aicra/response"
	"net/http"
	"path/filepath"
	"strings"
)

// Name returns the driver name
func (d *Import) Name() string { return "import" }

// Path returns the universal path from the source path
func (d Import) Path(_root, _folder, _src string) string {
	return strings.TrimSuffix(_src, ".go")
}

// Source returns the source path from the universal path
func (d Import) Source(_root, _folder, _path string) string {
	return fmt.Sprintf("%s.go", filepath.Join(_root, _folder, _path))

}

// Build returns the build path from the universal path
func (d Import) Build(_root, _folder, _path string) string {
	return filepath.Join(_root, _folder, _path)
}

// Compiled returns whether the driver has to be build
func (d Import) Compiled() bool { return false }

// RegisterController registers a new controller
func (d *Import) RegisterController(_path string, _controller Controller) error {

	// 1. init controllers if not already
	if d.Controllers == nil {
		d.Controllers = make(map[string]Controller)
	}

	// 2. fail if no controller
	if _controller == nil {
		return errors.New("nil controller")
	}

	// 3. Fail on invalid path
	if len(_path) < 1 || _path[0] != '/' {
		return errors.New("invalid controller path")
	}

	// 4. Store controller
	d.Controllers[_path] = _controller
	return nil

}

// RegisterMiddleware registers a new controller
func (d *Import) RegisterMiddlware(_path string, _middleware Middleware) error {

	// 1. init imddlewares if not already
	if d.Middlewares == nil {
		d.Middlewares = make(map[string]Middleware)
	}

	// 2. fail if no imddleware
	if _middleware == nil {
		return errors.New("nil imddleware")
	}

	// 3. Fail on invalid path
	if len(_path) < 1 || _path[0] != '/' {
		return errors.New("invalid imddleware path")
	}

	// 4. Store imddleware
	d.Middlewares[_path] = _middleware
	return nil

}

// RunController implements the Driver interface
func (d *Import) RunController(_path []string, _method string) (func(response.Arguments) response.Response, e.Error) {

	/* (1) Build controller path */
	path := strings.Join(_path, "-")

	/* (2) Check if controller exists */
	controller, ok := d.Controllers[path]
	if !ok {
		return nil, e.UncallableController
	}

	/* (3) Format method */
	method := strings.ToLower(_method)

	/* (4) Return method according to method */
	switch method {
	case "GET":
		return controller.Get, e.Success
	case "POST":
		return controller.Post, e.Success
	case "PUT":
		return controller.Put, e.Success
	case "DELETE":
		return controller.Delete, e.Success
	default:
		return nil, e.UncallableMethod
	}

	return nil, e.UncallableController
}

// LoadMiddleware returns a new middleware function; it must be a
// valid and existing folder/filename file with the .so extension
func (d *Import) LoadMiddleware(_path string) (func(http.Request, *[]string), error) {

	/* (1) Check plugin name */
	if len(_path) < 1 {
		return nil, errors.New("middleware name must not be empty")
	}

	/* (2) Check if middleware exists */
	middleware, ok := d.Middlewares[_path]
	if !ok {
		return nil, errors.New("middleware not found")
	}

	/* (3) Return middleware */
	return middleware.Inspect, nil
}
