package aicra

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/aicra/internal/dynfunc"
	"github.com/xdrm-io/aicra/validator"
)

// HandlerFunc defines the generic handler interface for services
type HandlerFunc[Req, Res any] func(context.Context, Req) (*Res, error)

const (
	// DefaultURILimit defines the default URI size to accept
	DefaultURILimit = 1024
	// DefaultBodyLimit defines the default body size to accept
	DefaultBodyLimit = 1024 * 1024 // 1MB
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

	// uriLimit is used to automatically reject requests that have a URI
	// exceeding `uriLimit` (in bytes). Negative values means there is no
	// limit. The default value (0) falls back to the default aicra limit
	uriLimit int
	// bodyLimit is used to automatically reject requests that have a body
	// exceeding `bodyLimit` (in bytes). Negative value means there is no
	// limit. The default value (0) falls back to the default aicra limit
	bodyLimit int64
}

// serviceHandler links a handler func to a service (method-path combination)
type serviceHandler struct {
	Method   string
	Path     string
	callable dynfunc.Callable
}

// SetURILimit defines the maximum size of request URIs that is accepted (in
// bytes)
func (b *Builder) SetURILimit(size int) {
	b.uriLimit = size
}

// SetBodyLimit defines the maximum size of request bodies that is accepted
// (in bytes)
func (b *Builder) SetBodyLimit(size int64) {
	b.bodyLimit = size
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
func Bind[Req, Res any](b *Builder, method, path string, fn HandlerFunc[Req, Res]) error {
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
		return fmt.Errorf("%s %q: %w", method, path, errUnknownService)
	}

	var callable, err = dynfunc.Build(service, dynfunc.HandlerFunc[Req, Res](fn))
	if err != nil {
		return fmt.Errorf("%s %q handler: %w", method, path, err)
	}

	b.handlers = append(b.handlers, &serviceHandler{
		Path:     path,
		Method:   method,
		callable: callable,
	})

	return nil
}

// Build a fully-featured HTTP server
func (b Builder) Build() (http.Handler, error) {
	if b.uriLimit == 0 {
		b.uriLimit = DefaultURILimit
	}
	if b.bodyLimit == 0 {
		b.bodyLimit = DefaultBodyLimit
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
			return nil, fmt.Errorf("%s %q: %w", service.Method, service.Pattern, errMissingHandler)
		}
	}

	return Handler(b), nil
}
