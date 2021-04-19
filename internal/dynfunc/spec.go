package dynfunc

import (
	"fmt"
	"reflect"
	"strings"

	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/internal/config"
)

type spec struct {
	Input  map[string]reflect.Type
	Output map[string]reflect.Type
}

// builds a spec from the configuration service
func makeSpec(service config.Service) *spec {
	s := &spec{
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

// checks for HandlerFn input arguments
func (s *spec) checkInput(impl reflect.Type, index int) error {
	var requiredInput, structIndex = index, index
	if len(s.Input) > 0 { // arguments struct
		requiredInput++
	}

	// missing arguments
	if impl.NumIn() > requiredInput {
		return errUnexpectedInput
	}

	// none required
	if len(s.Input) == 0 {
		return nil
	}

	// too much arguments
	if impl.NumIn() != requiredInput {
		return errMissingHandlerArgumentParam
	}

	// arg must be a struct
	structArg := impl.In(structIndex)
	if structArg.Kind() != reflect.Struct {
		return errMissingParamArgument
	}

	// check for invalid param
	for name, ptype := range s.Input {
		if name[0] == strings.ToLower(name)[0] {
			return fmt.Errorf("%s: %w", name, errUnexportedName)
		}

		field, exists := structArg.FieldByName(name)
		if !exists {
			return fmt.Errorf("%s: %w", name, errMissingParamFromConfig)
		}

		if !ptype.AssignableTo(field.Type) {
			return fmt.Errorf("%s: %w (%s instead of %s)", name, errWrongParamTypeFromConfig, field.Type, ptype)
		}
	}

	return nil
}

// checks for HandlerFn output arguments
func (s spec) checkOutput(impl reflect.Type) error {
	if impl.NumOut() < 1 {
		return errMissingHandlerOutput
	}

	// last output must be api.Err
	errOutput := impl.Out(impl.NumOut() - 1)
	if !errOutput.AssignableTo(reflect.TypeOf(api.ErrUnknown)) {
		return errMissingHandlerErrorOutput
	}

	// no output -> ok
	if len(s.Output) == 0 {
		return nil
	}

	if impl.NumOut() != 2 {
		return errMissingParamOutput
	}

	// fail if first output is not a pointer to struct
	structOutputPtr := impl.Out(0)
	if structOutputPtr.Kind() != reflect.Ptr {
		return errMissingParamOutput
	}

	structOutput := structOutputPtr.Elem()
	if structOutput.Kind() != reflect.Struct {
		return errMissingParamOutput
	}

	// fail on invalid output
	for name, ptype := range s.Output {
		if name[0] == strings.ToLower(name)[0] {
			return fmt.Errorf("%s: %w", name, errUnexportedName)
		}

		field, exists := structOutput.FieldByName(name)
		if !exists {
			return fmt.Errorf("%s: %w", name, errMissingOutputFromConfig)
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
