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
	tmp := []byte(strings.ToLower(_method))
	tmp[0] = tmp[0] - ('a' - 'A')
	method := string(tmp)

	/* (2) Try to load plugin */
	p, err2 := plugin.Open(path)
	if err2 != nil {
		return nil, err.UncallableController
	}

	/* (3) Try to extract method */
	m, err2 := p.Lookup(method)
	if err2 != nil {
		return nil, err.UncallableMethod
	}

	/* (4) Check signature */
	callable, validSignature := m.(func(response.Arguments) response.Response)
	if !validSignature {
		return nil, err.UncallableMethod
	}

	return callable, err.Success
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

	/* (4) Export wanted properties */
	inspect, err := p.Lookup("Inspect")
	if err != nil {
		return nil, fmt.Errorf("Missing method 'Inspect()'; %s", err)
	}

	/* (5) Cast Inspect */
	mware, ok := inspect.(func(http.Request, *[]string))
	if !ok {
		return nil, fmt.Errorf("Inspect() is malformed")
	}

	/* (6) Add type to registry */
	return mware, nil
}
