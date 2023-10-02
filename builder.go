package aicra

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/xdrm-io/aicra/internal/config"
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
	conf *config.API
	// user-defined handlers bound to endpoints from the configuration
	// Note: the index is the endpoint name
	handlers map[string]http.Handler
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

	// validators is the list of validators used to validate the configuration
	validators config.Validators
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

// With adds an http middleware on top of the http connection
//
// Authentication management can only be done with the WithContext() methods as
// the service associated with the request has not been found at this stage.
// This stage is perfect for logging or generic request management.
func (b *Builder) With(mw func(http.Handler) http.Handler) {
	if mw == nil {
		return
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
	if mw == nil {
		return
	}
	if b.ctxMiddlewares == nil {
		b.ctxMiddlewares = make([]func(http.Handler) http.Handler, 0)
	}
	b.ctxMiddlewares = append(b.ctxMiddlewares, mw)
}

// Setup the builder with its api definition file
// panics if already setup
func (b *Builder) Setup(r io.Reader) error {
	if b.conf != nil {
		return ErrAlreadySetup
	}
	b.conf = &config.API{}
	return json.NewDecoder(r).Decode(b.conf)
}

// Bind a handler to a REST service (method and pattern)
func (b *Builder) Bind(method, path string, fn http.HandlerFunc) error {
	if b.conf == nil || b.conf.Endpoints == nil {
		return ErrNotSetup
	}

	if fn == nil {
		return fmt.Errorf("'%s %s': %w", method, path, ErrNilHandler)
	}

	if b.handlers == nil {
		b.handlers = make(map[string]http.Handler, len(b.conf.Endpoints))
	}

	// find associated endpoint from config
	var endpoint *config.Endpoint
	for _, s := range b.conf.Endpoints {
		if method == s.Method && path == s.Pattern {
			endpoint = s
			break
		}
	}

	if endpoint == nil {
		return fmt.Errorf("'%s %s': %w", method, path, ErrUnknownService)
	}

	if _, ok := b.handlers[endpoint.Name]; ok {
		return fmt.Errorf("'%s %s': %w", method, path, ErrAlreadyBound)
	}
	b.handlers[endpoint.Name] = fn
	return nil
}

// Build a fully-featured HTTP server
func (b *Builder) Build(validators config.Validators) (http.Handler, error) {
	if b.conf == nil {
		return nil, ErrNotSetup
	}
	if validators == nil {
		return nil, ErrNilValidators
	}
	b.validators = validators

	if b.uriLimit == 0 {
		b.uriLimit = DefaultURILimit
	}
	if b.bodyLimit == 0 {
		b.bodyLimit = DefaultBodyLimit
	}

	for _, service := range b.conf.Endpoints {
		if _, ok := b.handlers[service.Name]; !ok {
			return nil, fmt.Errorf("%s %q: %w", service.Method, service.Pattern, ErrMissingHandler)
		}
	}

	if err := b.conf.RuntimeCheck(b.validators); err != nil {
		return nil, err
	}

	return Handler(*b), nil
}
