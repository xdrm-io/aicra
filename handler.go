package aicra

import (
	"fmt"
	"strings"

	"git.xdrm.io/go/aicra/dynamic"
	"git.xdrm.io/go/aicra/internal/config"
)

type handler struct {
	Method     string
	Path       string
	dynHandler *dynamic.Handler
}

// createHandler builds a handler from its http method and path
// also it checks whether the function signature is valid
func createHandler(method, path string, service config.Service, fn dynamic.HandlerFn) (*handler, error) {
	method = strings.ToUpper(method)

	dynHandler, err := dynamic.Build(fn, service)
	if err != nil {
		return nil, fmt.Errorf("%s '%s' handler: %w", method, path, err)
	}

	return &handler{
		Path:       path,
		Method:     method,
		dynHandler: dynHandler,
	}, nil
}
