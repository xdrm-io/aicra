package driver

import (
	"fmt"
	"path/filepath"
	"plugin"
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
	if _path == "" {
		return fmt.Sprintf("%s", filepath.Join(_root, ".build", _folder, "ROOT.so"))
	}
	return fmt.Sprintf("%s.so", filepath.Join(_root, ".build", _folder, _path))
}

// Compiled returns whether the driver has to be build
func (d Plugin) Compiled() bool { return true }

// LoadController returns a new Controller
func (d *Plugin) LoadController(_path string) (Controller, error) {

	/* 1. Try to load plugin */
	p, err := plugin.Open(_path)
	if err != nil {
		return nil, err
	}

	/* 2. Try to extract exported field */
	m, err := p.Lookup("Export")
	if err != nil {
		return nil, err
	}

	exporter, ok := m.(func() Controller)
	if !ok {
		return nil, err
	}

	/* 3. Controller */
	return exporter(), nil
}

// LoadMiddleware returns a new Middleware
func (d *Plugin) LoadMiddleware(_path string) (Middleware, error) {

	/* 1. Try to load plugin */
	p, err := plugin.Open(_path)
	if err != nil {
		return nil, err
	}

	/* 2. Try to extract exported field */
	m, err := p.Lookup("Export")
	if err != nil {
		return nil, err
	}

	exporter, ok := m.(func() Middleware)
	if !ok {
		return nil, err
	}

	return exporter(), nil
}

// LoadChecker returns a new Checker
func (d *Plugin) LoadChecker(_path string) (Checker, error) {

	/* 1. Try to load plugin */
	p, err := plugin.Open(_path)
	if err != nil {
		return nil, err
	}

	/* 2. Try to extract exported field */
	m, err := p.Lookup("Export")
	if err != nil {
		return nil, err
	}

	exporter, ok := m.(func() Checker)
	if !ok {
		return nil, err
	}

	return exporter(), nil
}
