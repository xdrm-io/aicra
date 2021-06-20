package dynfunc

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/internal/config"
)

// Handler represents a dynamic aicra service handler
type Handler struct {
	// signature defined from the service configuration
	signature *Signature
	// fn provided function that will be the service's handler
	fn interface{}
}

// Build a handler from a dynamic function and checks its signature against a
// service configuration
//e
// `fn` must have as a signature : `func(*api.Context, in) (*out, api.Err)`
//  - `in` is a struct{} containing a field for each service input (with valid reflect.Type)
//  - `out` is a struct{} containing a field for each service output (with valid reflect.Type)
//
// Special cases:
//  - it there is no input, `in` MUST be omitted
//  - it there is no output, `out` MUST be omitted
func Build(fn interface{}, service config.Service) (*Handler, error) {
	var (
		h = &Handler{
			signature: BuildSignature(service),
			fn:        fn,
		}
		fnType = reflect.TypeOf(fn)
	)

	if fnType.Kind() != reflect.Func {
		return nil, ErrHandlerNotFunc
	}
	if err := h.signature.ValidateInput(fnType); err != nil {
		return nil, fmt.Errorf("input: %w", err)
	}
	if err := h.signature.ValidateOutput(fnType); err != nil {
		return nil, fmt.Errorf("output: %w", err)
	}

	return h, nil
}

// Handle binds input `data` into the dynamic function and returns an output map
func (h *Handler) Handle(ctx context.Context, data map[string]interface{}) (map[string]interface{}, api.Err) {
	var (
		ert      = reflect.TypeOf(api.Err{})
		fnv      = reflect.ValueOf(h.fn)
		callArgs = make([]reflect.Value, 0)
	)

	// bind context
	callArgs = append(callArgs, reflect.ValueOf(ctx))

	inputStructRequired := fnv.Type().NumIn() > 1

	// bind input arguments
	if inputStructRequired {
		// create zero value struct
		var (
			callStructPtr = reflect.New(fnv.Type().In(1))
			callStruct    = callStructPtr.Elem()
		)

		// set each field
		for name := range h.signature.Input {
			field := callStruct.FieldByName(name)
			if !field.CanSet() {
				continue
			}

			// get value from @data
			value, provided := data[name]
			if !provided {
				continue
			}

			var refvalue = reflect.ValueOf(value)

			// T to pointer of T
			if field.Kind() == reflect.Ptr {
				var ptrType = field.Type().Elem()

				if !refvalue.Type().ConvertibleTo(ptrType) {
					log.Printf("Cannot convert %v into *%v", refvalue.Type(), ptrType)
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

			field.Set(refvalue.Convert(field.Type()))
		}
		callArgs = append(callArgs, callStruct)
	}

	// call the handler
	output := fnv.Call(callArgs)

	// no output OR pointer to output struct is nil
	outdata := make(map[string]interface{})
	if len(h.signature.Output) < 1 || output[0].IsNil() {
		var structerr = output[len(output)-1].Convert(ert)
		return outdata, api.Err{
			Code:   int(structerr.FieldByName("Code").Int()),
			Reason: structerr.FieldByName("Reason").String(),
			Status: int(structerr.FieldByName("Status").Int()),
		}
	}

	// extract struct from pointer
	returnStruct := output[0].Elem()

	for name := range h.signature.Output {
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
