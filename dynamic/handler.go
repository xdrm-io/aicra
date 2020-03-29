package dynamic

import (
	"fmt"
	"reflect"

	"git.xdrm.io/go/aicra/internal/config"
)

// Build a handler from a service configuration and a HandlerFn
//
// a HandlerFn must have as a signature : `func(api.Request, inputStruct) (outputStruct, api.Error)`
//  - `inputStruct` is a struct{} containing a field for each service input (with valid reflect.Type)
//  - `outputStruct` is a struct{} containing a field for each service output (with valid reflect.Type)
//
// Special cases:
//  - it there is no input, `inputStruct` can be omitted
//  - it there is no output, `outputStruct` can be omitted
func Build(fn HandlerFn, service config.Service) (*Handler, error) {
	h := &Handler{
		spec: makeSpec(service),
		fn:   fn,
	}

	fnv := reflect.ValueOf(fn)

	if fnv.Type().Kind() != reflect.Func {
		return nil, ErrHandlerNotFunc
	}

	if err := h.spec.checkInput(fnv); err != nil {
		return nil, fmt.Errorf("input: %w", err)
	}
	if err := h.spec.checkOutput(fnv); err != nil {
		return nil, fmt.Errorf("output: %w", err)
	}

	return h, nil
}

// Handle
func (h *Handler) Handle() {

}
