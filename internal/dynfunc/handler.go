package dynfunc

import (
	"fmt"
	"log"
	"reflect"

	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/internal/config"
)

// Handler represents a dynamic api handler
type Handler struct {
	spec *spec
	fn   interface{}
	// whether fn uses api.Ctx as 1st argument
	hasContext bool
}

// Build a handler from a service configuration and a dynamic function
//
// @fn must have as a signature : `func(inputStruct) (*outputStruct, api.Err)`
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

	impl := reflect.TypeOf(fn)

	if impl.Kind() != reflect.Func {
		return nil, errHandlerNotFunc
	}

	h.hasContext = impl.NumIn() >= 1 && reflect.TypeOf(api.Ctx{}).AssignableTo(impl.In(0))

	inputIndex := 0
	if h.hasContext {
		inputIndex = 1
	}

	if err := h.spec.checkInput(impl, inputIndex); err != nil {
		return nil, fmt.Errorf("input: %w", err)
	}
	if err := h.spec.checkOutput(impl); err != nil {
		return nil, fmt.Errorf("output: %w", err)
	}

	return h, nil
}

// Handle binds input @data into the dynamic function and returns map output
func (h *Handler) Handle(data map[string]interface{}) (map[string]interface{}, api.Err) {
	var ert = reflect.TypeOf(api.Err{})
	var fnv = reflect.ValueOf(h.fn)

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

			var refvalue = reflect.ValueOf(value)

			// T to pointer of T
			if field.Kind() == reflect.Ptr {
				var ptrType = field.Type().Elem()

				if !refvalue.Type().ConvertibleTo(ptrType) {
					log.Printf("Cannot convert %v into %v", refvalue.Type(), ptrType)
					return nil, api.ErrUncallableService
				}

				ptr := reflect.New(ptrType)
				ptr.Elem().Set(reflect.ValueOf(value).Convert(ptrType))

				field.Set(ptr)
				continue
			}
			if !reflect.ValueOf(value).Type().ConvertibleTo(field.Type()) {
				log.Printf("Cannot convert %v into %v", reflect.ValueOf(value).Type(), field.Type())
				return nil, api.ErrUncallableService
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
		var structerr = output[len(output)-1].Convert(ert)
		return outdata, api.Err{
			Code:   int(structerr.FieldByName("Code").Int()),
			Reason: structerr.FieldByName("Reason").String(),
			Status: int(structerr.FieldByName("Status").Int()),
		}
	}

	// extract struct from pointer
	returnStruct := output[0].Elem()

	for name := range h.spec.Output {
		field := returnStruct.FieldByName(name)
		outdata[name] = field.Interface()
	}

	// extract api.Err
	var structerr = output[len(output)-1].Convert(ert)
	return outdata, api.Err{
		Code:   int(structerr.FieldByName("Code").Int()),
		Reason: structerr.FieldByName("Reason").String(),
		Status: int(structerr.FieldByName("Status").Int()),
	}
}
