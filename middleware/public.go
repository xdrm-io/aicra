package middleware

import (
	"git.xdrm.io/go/aicra/driver"
	"net/http"
)

// CreateRegistry creates an empty registry
func CreateRegistry() Registry {
	return make(Registry)
}

// Add adds a new middleware for a path
func (reg Registry) Add(_path string, _element driver.Middleware) {
	reg[_path] = _element
}

// Run executes all middlewares (default browse order)
func (reg Registry) Run(req http.Request) []string {

	/* (1) Initialise scope */
	scope := make([]string, 0)

	/* (2) Execute each middleware */
	for _, mw := range reg {
		mw.Inspect(req, &scope)
	}

	return scope

}
