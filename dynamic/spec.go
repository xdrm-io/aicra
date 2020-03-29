package dynamic

import (
	"fmt"
	"reflect"

	"git.xdrm.io/go/aicra/api"
	"git.xdrm.io/go/aicra/internal/config"
)

// builds a spec from the configuration service
func makeSpec(service config.Service) spec {
	spec := spec{
		Input:  make(map[string]reflect.Type),
		Output: make(map[string]reflect.Type),
	}

	for _, param := range service.Input {
		// make a pointer if optional
		if param.Optional {
			spec.Input[param.Rename] = reflect.PtrTo(param.ExtractType)
			continue
		}
		spec.Input[param.Rename] = param.ExtractType
	}

	for _, param := range service.Output {
		// make a pointer if optional
		if param.Optional {
			spec.Output[param.Rename] = reflect.PtrTo(param.ExtractType)
			continue
		}
		spec.Output[param.Rename] = param.ExtractType
	}

	return spec
}

// checks for HandlerFn input arguments
func (s spec) checkInput(fnt reflect.Type) error {
	if fnt.NumIn() != 1 {
		return ErrMissingHandlerArgument
	}

	// arg[0] must be api.Request
	requestArg := fnt.In(0)
	if !requestArg.AssignableTo(reflect.TypeOf(api.Request{})) {
		return ErrMissingRequestArgument
	}

	// no input -> ok
	if len(s.Input) == 0 {
		return nil
	}

	if fnt.NumIn() < 2 {
		return ErrMissingHandlerArgumentParam
	}

	// arg[1] must be a struct
	structArg := fnt.In(1)
	if structArg.Kind() != reflect.Struct {
		return ErrMissingParamArgument
	}

	// check for invlaid param
	for name, ptype := range s.Input {
		field, exists := structArg.FieldByName(name)
		if !exists {
			return fmt.Errorf("%s: %w", name, ErrMissingParamFromConfig)
		}

		if !ptype.AssignableTo(field.Type) {
			return fmt.Errorf("%s: %w (%s)", name, ErrWrongParamTypeFromConfig, ptype)
		}
	}

	return nil
}

// checks for HandlerFn output arguments
func (s spec) checkOutput(fnt reflect.Type) error {
	if fnt.NumOut() < 1 {
		return ErrMissingHandlerOutput
	}

	// last output must be api.Error
	errOutput := fnt.Out(fnt.NumOut() - 1)
	if !errOutput.AssignableTo(reflect.TypeOf(api.ErrorUnknown)) {
		return ErrMissingHandlerErrorOutput
	}

	// no output -> ok
	if len(s.Output) == 0 {
		return nil
	}

	if fnt.NumOut() != 2 {
		return ErrMissingParamOutput
	}

	// fail if first output is not a struct
	structOutput := fnt.Out(0)
	if structOutput.Kind() != reflect.Struct {
		return ErrMissingParamArgument
	}

	// fail on invalid output
	for name, ptype := range s.Output {
		field, exists := structOutput.FieldByName(name)
		if !exists {
			return fmt.Errorf("%s: %w", name, ErrMissingOutputFromConfig)
		}

		if !ptype.AssignableTo(field.Type) {
			return fmt.Errorf("%s: %w (%s)", name, ErrWrongOutputTypeFromConfig, ptype)
		}
	}

	return nil
}
