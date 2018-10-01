package driver

import (
	"fmt"
	"git.xdrm.io/go/aicra/err"
	"git.xdrm.io/go/aicra/response"
	"net/http"
	"path/filepath"
	"plugin"
	"strings"
)

// Plugin tells the aicra instance to use the plugin driver to load controller/middleware executables
//
// It will load go .so plugins with the following interface :
//
// type Controller interface {
//		Get(d i.Arguments, r *i.Response) i.Response
//		Post(d i.Arguments, r *i.Response) i.Response
//		Put(d i.Arguments, r *i.Response) i.Response
//		Delete(d i.Arguments, r *i.Response) i.Response
// }
//
// The controllers are exported by calling the 'Export() Controller' method
type Plugin struct{}

// Name returns the driver name
func (d Plugin) Name() string { return "plugin" }

// Path returns the universal path from the source path
func (d Plugin) Path(_root, _folder, _src string) string {
	return filepath.Dir(_src)
}

// Source returns the source path from the universal path
func (d Plugin) Source(_root, _folder, _path string) string {
	return filepath.Join(_root, _folder, _path, "main.go")

}

// Build returns the build path from the universal path
func (d Plugin) Build(_root, _folder, _path string) string {
	return fmt.Sprintf("%s.so", filepath.Join(_root, ".build", _folder, _path))
}

// Compiled returns whether the driver has to be build
func (d Plugin) Compiled() bool { return true }

// RunController implements the Driver interface
func (d *Plugin) RunController(_path []string, _method string) (func(response.Arguments) response.Response, err.Error) {

	/* (1) Build controller path */
	path := strings.Join(_path, "-")
	if len(path) == 0 {
		path = fmt.Sprintf(".build/controller/ROOT.so")
	} else {
		path = fmt.Sprintf(".build/controller/%s.so", path)
	}

	/* (2) Format url */
	method := strings.ToLower(_method)

	/* (2) Try to load plugin */
	p, err2 := plugin.Open(path)
	if err2 != nil {
		return nil, err.UncallableController
	}

	/* (3) Try to extract exported field */
	m, err2 := p.Lookup("Export")
	if err2 != nil {
		return nil, err.UncallableController
	}

	exported, ok := m.(func() Controller)
	if !ok {
		return nil, err.UncallableController
	}

	/* (4) Controller */
	ctl := exported()

	/* (4) Check signature */
	switch method {
	case "get":
		return ctl.Get, err.Success
	case "post":
		return ctl.Post, err.Success
	case "put":
		return ctl.Put, err.Success
	case "delete":
		return ctl.Delete, err.Success
	}
	fmt.Printf("method: %s\n", method)

	return nil, err.UncallableMethod
}

// LoadMiddleware returns a new middleware function; it must be a
// valid and existing folder/filename file with the .so extension
func (d *Plugin) LoadMiddleware(_path string) (func(http.Request, *[]string), error) {

	// ignore non .so files
	if !strings.HasSuffix(_path, ".so") {
		return nil, fmt.Errorf("Invalid name")
	}

	/* (1) Check plugin name */
	if len(_path) < 1 {
		return nil, fmt.Errorf("Middleware name must not be empty")
	}

	/* (2) Check plugin extension */
	if !strings.HasSuffix(_path, ".so") {
		_path = fmt.Sprintf("%s.so", _path)
	}

	/* (3) Try to load the plugin */
	p, err := plugin.Open(_path)
	if err != nil {
		return nil, err
	}

	/* (4) Extract exported fields */
	mw, err := p.Lookup("Export")
	if err != nil {
		return nil, fmt.Errorf("Missing method 'Inspect()'; %s", err)
	}

	exported, ok := mw.(func() Middleware)
	if !ok {
		return nil, fmt.Errorf("Inspect() is malformed")
	}

	/* (5) Return Inspect method */
	return exported().Inspect, nil
}
