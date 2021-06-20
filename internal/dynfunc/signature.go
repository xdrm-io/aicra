package dynfunc

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/internal/config"
)

// Signature represents input/output arguments for service from the aicra configuration
type Signature struct {
	// Input arguments of the service
	Input map[string]reflect.Type
	// Output arguments of the service
	Output map[string]reflect.Type
}

// BuildSignature builds a signature for a service configuration
func BuildSignature(service config.Service) *Signature {
	s := &Signature{
		Input:  make(map[string]reflect.Type),
		Output: make(map[string]reflect.Type),
	}

	for _, param := range service.Input {
		if len(param.Rename) < 1 {
			continue
		}
		// make a pointer if optional
		if param.Optional {
			s.Input[param.Rename] = reflect.PtrTo(param.ExtractType)
			continue
		}
		s.Input[param.Rename] = param.ExtractType
	}

	for _, param := range service.Output {
		if len(param.Rename) < 1 {
			continue
		}
		s.Output[param.Rename] = param.ExtractType
	}

	return s
}

// ValidateInput validates a handler's input arguments against the service signature
func (s *Signature) ValidateInput(handlerType reflect.Type) error {
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()

	// missing or invalid first arg: context.Context
	if handlerType.NumIn() < 1 {
		return errMissingHandlerContextArgument
	}
	firstArgType := handlerType.In(0)

	if !firstArgType.Implements(ctxType) {
		return fmt.Errorf("fock")
	}

	// no input required
	if len(s.Input) == 0 {
		// input struct provided
		if handlerType.NumIn() > 1 {
			return errUnexpectedInput
		}
		return nil
	}

	// too much arguments
	if handlerType.NumIn() > 2 {
		return errMissingHandlerInputArgument
	}

	// arg must be a struct
	inStruct := handlerType.In(1)
	if inStruct.Kind() != reflect.Struct {
		return errMissingParamArgument
	}

	// check for invalid param
	for name, ptype := range s.Input {
		if name[0] == strings.ToLower(name)[0] {
			return fmt.Errorf("%s: %w", name, errUnexportedName)
		}

		field, exists := inStruct.FieldByName(name)
		if !exists {
			return fmt.Errorf("%s: %w", name, errMissingConfigArgument)
		}

		if !ptype.AssignableTo(field.Type) {
			return fmt.Errorf("%s: %w (%s instead of %s)", name, errWrongParamTypeFromConfig, field.Type, ptype)
		}
	}

	return nil
}

// ValidateOutput validates a handler's output arguments against the service signature
func (s Signature) ValidateOutput(handlerType reflect.Type) error {
	errType := reflect.TypeOf(api.ErrUnknown)

	if handlerType.NumOut() < 1 {
		return errMissingHandlerErrorArgument
	}

	// last output must be api.Err
	lastArgType := handlerType.Out(handlerType.NumOut() - 1)
	if !lastArgType.AssignableTo(errType) {
		return errMissingHandlerErrorArgument
	}

	// no output -> ok
	if len(s.Output) == 0 {
		return nil
	}

	if handlerType.NumOut() < 2 {
		return errMissingHandlerOutputArgument
	}

	// fail if first output is not a pointer to struct
	outStructPtr := handlerType.Out(0)
	if outStructPtr.Kind() != reflect.Ptr {
		return errWrongOutputArgumentType
	}

	outStruct := outStructPtr.Elem()
	if outStruct.Kind() != reflect.Struct {
		return errWrongOutputArgumentType
	}

	// fail on invalid output
	for name, ptype := range s.Output {
		if name[0] == strings.ToLower(name)[0] {
			return fmt.Errorf("%s: %w", name, errUnexportedName)
		}

		field, exists := outStruct.FieldByName(name)
		if !exists {
			return fmt.Errorf("%s: %w", name, errMissingConfigArgument)
		}

		// ignore types evalutating to nil
		if ptype == nil {
			continue
		}

		if !field.Type.ConvertibleTo(ptype) {
			return fmt.Errorf("%s: %w (%s instead of %s)", name, errWrongParamTypeFromConfig, field.Type, ptype)
		}
	}

	return nil
}
