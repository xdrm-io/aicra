package dynfunc

import (
	"context"
	"fmt"
	"reflect"

	"github.com/xdrm-io/aicra/internal/config"
)

// HandlerFunc represents an user-provdided generic handler
type HandlerFunc[Req, Res any] func(context.Context, Req) (*Res, error)

// Callable usually wraps a HandlerFn but has a common signature
type Callable func(context.Context, map[string]interface{}) (map[string]interface{}, error)

// Build a dynamic handler from a generic HandlerFn . Fails when the function
// does not match the expected service signature (input and output arguments)
// according to the configuration.
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
//  - when no input is configured, the `in` struct MUST be empty
//  - when no output is configured, the `out` struct MUST be empty
func Build[Req, Res any](service *config.Service, fn HandlerFunc[Req, Res]) (Callable, error) {
	var signature = NewSignature(service)

	var (
		treq = reflect.TypeOf((*Req)(nil)).Elem()
		tres = reflect.TypeOf((*Res)(nil)).Elem()
	)

	if err := signature.ValidateRequest(treq); err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	if err := signature.ValidateResponse(tres); err != nil {
		return nil, fmt.Errorf("response: %w", err)
	}

	return Wrap(signature, fn), nil
}

// Wrap a generic handler into a callable function
func Wrap[Req, Res any](s *Signature, fn HandlerFunc[Req, Res]) Callable {
	// preprocess indexes to avoid using FieldByName()
	var (
		treq = reflect.TypeOf((*Req)(nil)).Elem()
		tres = reflect.TypeOf((*Res)(nil)).Elem()
	)

	var reqIndex = make(map[string][]int, len(s.In))
	for name := range s.In {
		if field, ok := treq.FieldByName(name); ok {
			reqIndex[name] = make([]int, len(field.Index))
			copy(reqIndex[name], field.Index)
		}
	}

	var resIndex = make(map[string][]int, len(s.Out))
	for name := range s.Out {
		if field, ok := tres.FieldByName(name); ok {
			resIndex[name] = make([]int, len(field.Index))
			copy(resIndex[name], field.Index)
		}
	}

	return func(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
		var (
			tfn       = reflect.TypeOf(fn)
			hasOutput = len(s.Out) > 0
		)

		// create zero value struct
		var (
			inStructPtr = reflect.New(tfn.In(1))
			inStruct    = inStructPtr.Elem()
		)

		// convert map[string]interface{} into Req
		for name := range s.In {
			field := inStruct.FieldByIndex(reqIndex[name])

			// get value from @data
			value, provided := in[name]
			if !provided {
				continue
			}

			vvalue := reflect.ValueOf(value)
			tvalue := reflect.TypeOf(value)

			// convert T to pointer of T
			if field.Kind() == reflect.Ptr {
				var tPtr = field.Type().Elem()
				if !tvalue.ConvertibleTo(tPtr) {
					panic(fmt.Errorf("cannot convert %v into *%v", tvalue, tPtr))
				}

				ptr := reflect.New(tPtr)
				ptr.Elem().Set(vvalue.Convert(tPtr))
				field.Set(ptr)
				continue
			}

			// not convertible
			if !tvalue.ConvertibleTo(field.Type()) {
				panic(fmt.Errorf("cannot convert %v into %v", tvalue, field.Type()))
			}

			// non-pointer values
			field.Set(vvalue.Convert(field.Type()))
		}

		// call the handler
		var (
			req      = inStruct.Interface().(Req)
			res, err = fn(ctx, req)
			vres     = reflect.ValueOf(res).Elem()
		)

		// no output OR pointer to output struct is nil
		if !hasOutput || res == nil {
			return nil, err
		}

		// convert Res to map[string]interface{}
		out := make(map[string]interface{}, len(s.Out))
		for name := range s.Out {
			field := vres.FieldByIndex(resIndex[name])
			out[name] = field.Interface()
		}
		return out, err
	}
}
