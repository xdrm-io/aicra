package dynfunc

import (
	"context"
	"fmt"
	"reflect"

	"github.com/xdrm-io/aicra/internal/config"
)

// Handler represents a dynamic aicra service handler
type Handler struct {
	// signature defined from the service configuration
	signature *Signature
	// fn provided function that will be the service's handler
	fn interface{}
}

// Build a dynamic handler from a generic function (interface{}). Fail when the
// function does not match the expected service signature (input and output
// arguments) according to the configuration.
//
// `fn` must have as a signature : `func(context.Context, in) (*out, api.Err)`
//  - `in`  is a struct{} containing a field for each service input
//  - `out` is a struct{} containing a field for each service output
//
// Struct field names must be literally the same as the "name" field from the
// configuration, or the argument key if no "name" is provided.
//
// Input struct field types must match the associated validator GoType().
// Optional input arguments must be pointers to the validator's GoType().
// Output struct field types must match output types.
//
// Special cases:
//  - when no input is configured, the `in` struct MUST be dropped
//  - when no output is configured, the `out` struct MUST be dropped
func Build(fn interface{}, service config.Service) (*Handler, error) {
	var (
		h = &Handler{
			signature: FromConfig(service),
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
func (h *Handler) Handle(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	var (
		vFn   = reflect.ValueOf(h.fn)
		tFn   = reflect.TypeOf(h.fn)
		input = make([]reflect.Value, 0)

		hasInput  = len(h.signature.In) > 0
		hasOutput = len(h.signature.Out) > 0
	)

	// bind context
	input = append(input, reflect.ValueOf(ctx))

	// bind input arguments
	if hasInput {
		// create zero value struct
		var (
			inStructPtr = reflect.New(tFn.In(1))
			inStruct    = inStructPtr.Elem()
		)

		// set each field
		for name := range h.signature.In {
			field := inStruct.FieldByName(name)
			if !field.CanSet() {
				panic(fmt.Errorf("cannot set field %q", name))
			}

			// get value from @data
			value, provided := data[name]
			if !provided {
				continue
			}

			vValue := reflect.ValueOf(value)
			tValue := reflect.TypeOf(value)

			// convert T to pointer of T
			if field.Kind() == reflect.Ptr {
				var tPtr = field.Type().Elem()
				if !tValue.ConvertibleTo(tPtr) {
					panic(fmt.Errorf("cannot convert %v into *%v", tValue, tPtr))
				}

				ptr := reflect.New(tPtr)
				ptr.Elem().Set(vValue.Convert(tPtr))
				field.Set(ptr)
				continue
			}

			// not convertible
			if !tValue.ConvertibleTo(field.Type()) {
				panic(fmt.Errorf("cannot convert %v into %v", tValue, field.Type()))
			}

			// non-pointer values
			field.Set(vValue.Convert(field.Type()))
		}
		input = append(input, inStruct)
	}

	// call the handler
	output := vFn.Call(input)

	var err error
	if !output[len(output)-1].IsNil() {
		err = output[len(output)-1].Interface().(error)
	}

	// no output OR pointer to output struct is nil
	outdata := make(map[string]interface{})
	if !hasOutput || output[0].IsNil() {
		return outdata, err
	}

	// extract struct from pointer
	outStruct := output[0].Elem()
	for name := range h.signature.Out {
		field := outStruct.FieldByName(name)
		outdata[name] = field.Interface()
	}
	return outdata, err
}
