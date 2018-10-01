package driver

import (
	"path/filepath"
)

// Generic tells the aicra instance to use the generic driver to load controller/middleware executables
//
// It will call an executable with the json input into the standard input (argument 1)
//    the HTTP method is send as the key _HTTP_METHOD_ (in upper case)
// The standard output must be a json corresponding to the data
type Generic struct{}

// Name returns the driver name
func (d *Generic) Name() string { return "generic" }

// Path returns the universal path from the source path
func (d Generic) Path(_root, _folder, _src string) string {
	return _src
}

// Source returns the source path from the universal path
func (d Generic) Source(_root, _folder, _path string) string {
	return filepath.Join(_root, _folder, _path)

}

// Build returns the build path from the universal path
func (d Generic) Build(_root, _folder, _path string) string {
	return filepath.Join(_root, _folder, _path)
}

// Compiled returns whether the driver has to be build
func (d Generic) Compiled() bool { return false }

// LoadController implements the Driver interface
func (d *Generic) LoadController(_path string) (Controller, error) {
	return genericController(_path), nil
}

// LoadMiddleware returns a new middleware; it must be a
// valid and existing folder/filename file
func (d *Generic) LoadMiddleware(_path string) (Middleware, error) {
	return genericMiddleware(_path), nil
}

// LoadChecker returns a new middleware; it must be a
// valid and existing folder/filename file
func (d *Generic) LoadChecker(_path string) (Checker, error) {
	return genericChecker(_path), nil
}
