package dynfunc

import (
	"fmt"
	"reflect"

	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/internal/config"
)

// Handler represents a dynamic api handler
type Handler struct {
	spec spec
	fn   interface{}
}

// Build a handler from a service configuration and a dynamic function
//
// @fn must have as a signature : `func(inputStruct) (*outputStruct, api.Error)`
//  - `inputStruct` is a struct{} containing a field for each service input (with valid reflect.Type)
//  - `outputStruct` is a struct{} containing a field for each service output (with valid reflect.Type)
//
// Special cases:
//  - it there is no input, `inputStruct` must be omitted
//  - it there is no output, `outputStruct` must be omitted
func Build(fn interface{}, service config.Service) (*Handler, error) {
	h := &Handler{
		spec: makeSpec(service),
		fn:   fn,
	}

	fnv := reflect.ValueOf(fn)

	if fnv.Type().Kind() != reflect.Func {
		return nil, errHandlerNotFunc
	}

	if err := h.spec.checkInput(fnv); err != nil {
		return nil, fmt.Errorf("input: %w", err)
	}
	if err := h.spec.checkOutput(fnv); err != nil {
		return nil, fmt.Errorf("output: %w", err)
	}

	return h, nil
}

// Handle binds input @data into the dynamic function and returns map output
func (h *Handler) Handle(data map[string]interface{}) (map[string]interface{}, api.Error) {
	fnv := reflect.ValueOf(h.fn)

	callArgs := []reflect.Value{}

	// bind input data
	if fnv.Type().NumIn() > 0 {
		// create zero value struct
		callStructPtr := reflect.New(fnv.Type().In(0))
		callStruct := callStructPtr.Elem()

		// set each field
		for name := range h.spec.Input {
			field := callStruct.FieldByName(name)
			if !field.CanSet() {
				continue
			}

			// get value from @data
			value, inData := data[name]
			if !inData {
				continue
			}
			field.Set(reflect.ValueOf(value).Convert(field.Type()))
		}
		callArgs = append(callArgs, callStruct)
	}

	// call the HandlerFn
	output := fnv.Call(callArgs)

	// no output OR pointer to output struct is nil
	outdata := make(map[string]interface{})
	if len(h.spec.Output) < 1 || output[0].IsNil() {
		return outdata, api.Error(output[len(output)-1].Int())
	}

	// extract struct from pointer
	returnStruct := output[0].Elem()

	for name := range h.spec.Output {
		field := returnStruct.FieldByName(name)
		outdata[name] = field.Interface()
	}

	// extract api.Error
	return outdata, api.Error(output[len(output)-1].Int())
}
