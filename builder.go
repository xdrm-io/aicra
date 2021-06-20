package aicra

import (
	"fmt"
	"io"
	"net/http"

	"git.xdrm.io/go/aicra/datatype"
	"git.xdrm.io/go/aicra/internal/config"
	"git.xdrm.io/go/aicra/internal/dynfunc"
)

// Builder for an aicra server
type Builder struct {
	// the server configuration defining available services
	conf *config.Server
	// user-defined handlers bound to services from the configuration
	handlers []*apiHandler
	// http middlewares wrapping the entire http connection (e.g. logger)
	middlewares []func(http.Handler) http.Handler
	// custom middlewares only wrapping the service handler of a request
	// they will benefit from the request's context that contains service-specific
	// information (e.g. required permisisons from the configuration)
	ctxMiddlewares []func(http.Handler) http.Handler
}

// represents an api handler (method-pattern combination)
type apiHandler struct {
	Method string
	Path   string
	dyn    *dynfunc.Handler
}

// AddType adds an available datatype to the api definition
func (b *Builder) AddType(t datatype.T) error {
	if b.conf == nil {
		b.conf = &config.Server{}
	}
	if b.conf.Services != nil {
		return errLateType
	}
	if b.conf.Types == nil {
		b.conf.Types = make([]datatype.T, 0)
	}
	b.conf.Types = append(b.conf.Types, t)
	return nil
}

// With adds an http middleware on top of the http connection
//
// Authentication management can only be done with the WithContext() methods as
// the service associated with the request has not been found at this stage.
// This stage is perfect for logging or generic request management.
func (b *Builder) With(mw func(http.Handler) http.Handler) {
	if b.conf == nil {
		b.conf = &config.Server{}
	}
	if b.middlewares == nil {
		b.middlewares = make([]func(http.Handler) http.Handler, 0)
	}
	b.middlewares = append(b.middlewares, mw)
}

// WithContext adds an http middleware with the fully loaded context
//
// Logging or generic request management should be done with the With() method as
// it wraps the full http connection. Middlewares added through this method only
// wrap the user-defined service handler. The context.Context is filled with useful
// data that can be access with api.GetRequest(), api.GetResponseWriter(),
// api.GetAuth(), etc methods.
func (b *Builder) WithContext(mw func(http.Handler) http.Handler) {
	if b.conf == nil {
		b.conf = &config.Server{}
	}
	if b.ctxMiddlewares == nil {
		b.ctxMiddlewares = make([]func(http.Handler) http.Handler, 0)
	}
	b.ctxMiddlewares = append(b.ctxMiddlewares, mw)
}

// Setup the builder with its api definition file
// panics if already setup
func (b *Builder) Setup(r io.Reader) error {
	if b.conf == nil {
		b.conf = &config.Server{}
	}
	if b.conf.Services != nil {
		panic(errAlreadySetup)
	}
	return b.conf.Parse(r)
}

// Bind a dynamic handler to a REST service (method and pattern)
func (b *Builder) Bind(method, path string, fn interface{}) error {
	if b.conf.Services == nil {
		return errNotSetup
	}

	// find associated service from config
	var service *config.Service
	for _, s := range b.conf.Services {
		if method == s.Method && path == s.Pattern {
			service = s
			break
		}
	}

	if service == nil {
		return fmt.Errorf("%s '%s': %w", method, path, errUnknownService)
	}

	var dyn, err = dynfunc.Build(fn, *service)
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

// Get is equivalent to Bind(http.MethodGet)
func (b *Builder) Get(path string, fn interface{}) error {
	return b.Bind(http.MethodGet, path, fn)
}

// Post is equivalent to Bind(http.MethodPost)
func (b *Builder) Post(path string, fn interface{}) error {
	return b.Bind(http.MethodPost, path, fn)
}

// Put is equivalent to Bind(http.MethodPut)
func (b *Builder) Put(path string, fn interface{}) error {
	return b.Bind(http.MethodPut, path, fn)
}

// Delete is equivalent to Bind(http.MethodDelete)
func (b *Builder) Delete(path string, fn interface{}) error {
	return b.Bind(http.MethodDelete, path, fn)
}

// Build a fully-featured HTTP server
func (b Builder) Build() (http.Handler, error) {

	for _, service := range b.conf.Services {
		var isHandled bool
		for _, handler := range b.handlers {
			if handler.Method == service.Method && handler.Path == service.Pattern {
				isHandled = true
				break
			}
		}
		if !isHandled {
			return nil, fmt.Errorf("%s '%s': %w", service.Method, service.Pattern, errMissingHandler)
		}
	}

	return Handler(b), nil
}
