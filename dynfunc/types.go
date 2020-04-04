package dynfunc

import "reflect"

// Handler represents a dynamic api handler
type Handler struct {
	spec spec
	fn   interface{}
}

type spec struct {
	Input  map[string]reflect.Type
	Output map[string]reflect.Type
}
