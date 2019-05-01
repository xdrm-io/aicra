package api

import (
	"strings"
)

// HandlerFunc manages an API request
type HandlerFunc func(Request, *Response)

// Handler is an API handler ready to be bound
type Handler struct {
	path   string
	method string
	handle HandlerFunc
}

// NewHandler builds a handler from its http method and path
func NewHandler(method, path string, handlerFunc HandlerFunc) *Handler {
	return &Handler{
		path:   path,
		method: strings.ToUpper(method),
		handle: handlerFunc,
	}
}

// Handle fires a handler
func (h *Handler) Handle(req Request, res *Response) {
	h.handle(req, res)
}

// GetMethod returns the handler's HTTP method
func (h *Handler) GetMethod() string {
	return h.method
}

// GetPath returns the handler's path
func (h *Handler) GetPath() string {
	return h.path
}
