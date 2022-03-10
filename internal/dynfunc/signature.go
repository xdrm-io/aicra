package dynfunc

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/xdrm-io/aicra/internal/config"
)

// Signature represents input and output arguments for a specific service of the
// aicra configuration
type Signature struct {
	In  map[string]reflect.Type
	Out map[string]reflect.Type
}

// FromConfig builds the handler signature type from a service's configuration
func FromConfig(service config.Service) *Signature {
	s := &Signature{
		In:  make(map[string]reflect.Type),
		Out: make(map[string]reflect.Type),
	}

	for _, param := range service.Input {
		if len(param.Rename) < 1 {
			continue
		}
		// make a pointer if optional
		if param.Optional {
			s.In[param.Rename] = reflect.PtrTo(param.GoType)
			continue
		}
		s.In[param.Rename] = param.GoType
	}

	for _, param := range service.Output {
		if len(param.Rename) < 1 {
			continue
		}
		s.Out[param.Rename] = param.GoType
	}
	return s
}

// ValidateInput arguments of a handler against the signature
func (s *Signature) ValidateInput(tHandler reflect.Type) error {
	tContext := reflect.TypeOf((*context.Context)(nil)).Elem()

	// context.Context first argument missing/invalid
	if tHandler.NumIn() < 1 {
		return ErrMissingHandlerContextArgument
	}
	tFirst := tHandler.In(0)
	if !tFirst.Implements(tContext) {
		return ErrInvalidHandlerContextArgument
	}

	// no input required
	if len(s.In) == 0 {
		// fail when input struct is still provided
		if tHandler.NumIn() > 1 {
			return ErrUnexpectedInput
		}
		return nil
	}

	// too much arguments
	if tHandler.NumIn() != 2 {
		return ErrMissingHandlerInputArgument
	}

	// must be a struct
	tInStruct := tHandler.In(1)
	if tInStruct.Kind() != reflect.Struct {
		return ErrMissingParamArgument
	}

	// check for invalid param
	for name, tParam := range s.In {
		if name[0] == strings.ToLower(name)[0] {
			return fmt.Errorf("%s: %w", name, ErrUnexportedName)
		}

		field, exists := tInStruct.FieldByName(name)
		if !exists {
			return fmt.Errorf("%s: %w", name, ErrMissingConfigArgument)
		}

		if !tParam.AssignableTo(field.Type) {
			return fmt.Errorf("%s: %w (%s instead of %s)", name, ErrWrongParamTypeFromConfig, field.Type, tParam)
		}
	}
	return nil
}

// ValidateOutput arguments of a handler against the signature
func (s Signature) ValidateOutput(tHandler reflect.Type) error {
	tError := reflect.TypeOf((*error)(nil)).Elem()

	if tHandler.NumOut() < 1 {
		return ErrMissingHandlerErrorArgument
	}

	// last argument must be an error
	tLast := tHandler.Out(tHandler.NumOut() - 1)
	if !tLast.AssignableTo(tError) {
		return ErrInvalidHandlerErrorArgument
	}

	// no output required
	if len(s.Out) == 0 {
		// fail when output struct is still provided
		if tHandler.NumOut() > 1 {
			return ErrUnexpectedOutput
		}
		return nil
	}

	if tHandler.NumOut() < 2 {
		return ErrMissingHandlerOutputArgument
	}

	// fail if first output is not a pointer to struct
	tOutStructPtr := tHandler.Out(0)
	if tOutStructPtr.Kind() != reflect.Ptr {
		return ErrWrongOutputArgumentType
	}

	tOutStruct := tOutStructPtr.Elem()
	if tOutStruct.Kind() != reflect.Struct {
		return ErrWrongOutputArgumentType
	}

	// fail on invalid output
	for name, tParam := range s.Out {
		if name[0] == strings.ToLower(name)[0] {
			return fmt.Errorf("%s: %w", name, ErrUnexportedName)
		}

		field, exists := tOutStruct.FieldByName(name)
		if !exists {
			return fmt.Errorf("%s: %w", name, ErrMissingConfigArgument)
		}

		// ignore types evalutating to nil
		if tParam == nil {
			continue
		}

		if !field.Type.ConvertibleTo(tParam) {
			return fmt.Errorf("%s: %w (%s instead of %s)", name, ErrWrongParamTypeFromConfig, field.Type, tParam)
		}
	}

	return nil
}
