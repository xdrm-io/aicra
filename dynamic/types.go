package dynamic

import "reflect"

// HandlerFn defines a dynamic handler function
type HandlerFn interface{}

// Handler represents a dynamic api handler
type Handler struct {
	spec spec
	fn   HandlerFn
}

type spec struct {
	Input  map[string]reflect.Type
	Output map[string]reflect.Type
}
