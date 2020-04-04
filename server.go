package aicra

import (
	"fmt"
	"io"
	"net/http"

	"git.xdrm.io/go/aicra/datatype"
	"git.xdrm.io/go/aicra/dynfunc"
	"git.xdrm.io/go/aicra/internal/config"
)

// Builder for an aicra server
type Builder struct {
	conf     *config.Server
	handlers []*apiHandler
}

// represents an server handler
type apiHandler struct {
	Method string
	Path   string
	dyn    *dynfunc.Handler
}

// AddType adds an available datatype to the api definition
func (b *Builder) AddType(t datatype.T) {
	if b.conf == nil {
		b.conf = &config.Server{}
	}
	if b.conf.Services != nil {
		panic(ErrLateType)
	}
	b.conf.Types = append(b.conf.Types, t)
}

// Setup the builder with its api definition
// panics if already setup
func (b *Builder) Setup(r io.Reader) error {
	if b.conf == nil {
		b.conf = &config.Server{}
	}
	if b.conf.Services != nil {
		panic(ErrAlreadySetup)
	}
	return b.conf.Parse(r)
}

// Bind a dynamic handler to a REST service
func (b *Builder) Bind(method, path string, fn interface{}) error {
	if b.conf.Services == nil {
		return ErrNotSetup
	}

	// find associated service
	var service *config.Service
	for _, s := range b.conf.Services {
		if method == s.Method && path == s.Pattern {
			service = s
			break
		}
	}

	if service == nil {
		return fmt.Errorf("%s '%s': %w", method, path, ErrUnknownService)
	}

	dyn, err := dynfunc.Build(fn, *service)
	if err != nil {
		return fmt.Errorf("%s '%s' handler: %w", method, path, err)
	}

	b.handlers = append(b.handlers, &apiHandler{
		Path:   path,
		Method: method,
		dyn:    dyn,
	})

	return nil
}

// Build a fully-featured HTTP server
func (b Builder) Build() (http.Handler, error) {

	for _, service := range b.conf.Services {
		var hasAssociatedHandler bool
		for _, handler := range b.handlers {
			if handler.Method == service.Method && handler.Path == service.Pattern {
				hasAssociatedHandler = true
				break
			}
		}
		if !hasAssociatedHandler {
			return nil, fmt.Errorf("%s '%s': %w", service.Method, service.Pattern, ErrMissingHandler)
		}
	}

	return httpHandler(b), nil
}
