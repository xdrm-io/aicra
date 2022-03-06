package aicra

import (
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/aicra/internal/dynfunc"
	"github.com/xdrm-io/aicra/validator"
)

const (
	// DefaultMaxURISize defines the default URI size to accept
	DefaultMaxURISize = 1024
	// DefaultMaxBodySize defines the default body size to accept
	DefaultMaxBodySize = 1024 * 1024 // 1MB
)

// Builder for an aicra server
type Builder struct {
	// the server configuration defining available services
	conf *config.Server
	// respond func defines how to write data and error into an http response,
	// defaults to `DefaultResponder`.
	respond Responder
	// user-defined handlers bound to services from the configuration
	handlers []*serviceHandler
	// http middlewares wrapping the entire http connection (e.g. logger)
	middlewares []func(http.Handler) http.Handler
	// custom middlewares only wrapping the service handler of a request
	// they will benefit from the request's context that contains service-specific
	// information (e.g. required permissions from the configuration)
	ctxMiddlewares []func(http.Handler) http.Handler

	// maxURISize is used to automatically reject requests that have a URI
	// exceeding `maxURISize` (in bytes). Negative values means there is no
	// limit. The default value (0) falls back to the default aicra limit
	maxURISize int
	// maxBodySize is used to automatically reject requests that have a body
	// exceeding `maxBodySize` (in bytes). Negative value means there is no
	// limit. The default value (0) falls back to the default aicra limit
	maxBodySize int64
}

// serviceHandler links a handler func to a service (method-path combination)
type serviceHandler struct {
	Method string
	Path   string
	dyn    *dynfunc.Handler
}

// SetMaxURISize defines the maximum size of request URIs that is accepted (in
// bytes)
func (b *Builder) SetMaxURISize(size int) {
	b.maxURISize = size
}

// SetMaxBodySize defines the maximum size of request bodies that is accepted
// (in bytes)
func (b *Builder) SetMaxBodySize(size int64) {
	b.maxBodySize = size
}

// Input adds an available validator for input arguments
func (b *Builder) Input(t validator.Type) error {
	if b.conf == nil {
		b.conf = &config.Server{}
	}
	if b.conf.Services != nil {
		return errLateType
	}
	b.conf.AddInputValidator(t)
	return nil
}

// Output adds an type available for output arguments as well as a value example.
// Some examples:
// - Output("uint",  uint(0))
// - Output("user",  model.User{})
// - Output("users", []model.User{})
func (b *Builder) Output(name string, sample interface{}) error {
	if b.conf == nil {
		b.conf = &config.Server{}
	}
	if b.conf.Services != nil {
		return errLateType
	}
	b.conf.AddOutputValidator(name, reflect.TypeOf(sample))
	return nil
}

// RespondWith defines the server responder, i.e. how to write data and error
// into the http response.
func (b *Builder) RespondWith(responder Responder) error {
	if responder == nil {
		return errNilResponder
	}
	b.respond = responder
	return nil
}

// With adds an http middleware on top of the http connection
//
// Authentication management can only be done with the WithContext() methods as
// the service associated with the request has not been found at this stage.
// This stage is perfect for logging or generic request management.
func (b *Builder) With(mw func(http.Handler) http.Handler) {
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
		return errAlreadySetup
	}
	return b.conf.Parse(r)
}

// Bind a dynamic handler to a REST service (method and pattern)
func (b *Builder) Bind(method, path string, fn interface{}) error {
	if b.conf == nil || b.conf.Services == nil {
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

	b.handlers = append(b.handlers, &serviceHandler{
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
	if b.maxURISize == 0 {
		b.maxURISize = DefaultMaxURISize
	}
	if b.maxBodySize == 0 {
		b.maxBodySize = DefaultMaxBodySize
	}

	if b.respond == nil {
		b.respond = DefaultResponder
	}

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
