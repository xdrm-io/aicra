package api

import (
	"strings"
)

// HandlerFn defines the handler signature
type HandlerFn func(req Request, res *Response) Error

// Handler is an API handler ready to be bound
type Handler struct {
	path   string
	method string
	Fn     HandlerFn
}

// NewHandler builds a handler from its http method and path
func NewHandler(method, path string, fn HandlerFn) (*Handler, error) {
	return &Handler{
		path:   path,
		method: strings.ToUpper(method),
		Fn:     fn,
	}, nil
}

// GetMethod returns the handler's HTTP method
func (h *Handler) GetMethod() string {
	return h.method
}

// GetPath returns the handler's path
func (h *Handler) GetPath() string {
	return h.path
}
