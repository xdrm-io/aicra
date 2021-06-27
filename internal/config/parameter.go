package config

import (
	"reflect"

	"github.com/xdrm-io/aicra/validator"
)

// Parameter represents a parameter definition (from api.json)
type Parameter struct {
	Description string `json:"info"`
	Type        string `json:"type"`
	Rename      string `json:"name,omitempty"`
	Optional    bool
	// GoType is the type the Validator will cast into
	GoType reflect.Type
	// Validator is inferred from the "type" property
	Validator validator.ValidateFunc
}

func (param *Parameter) validate(validators ...validator.Type) error {
	if len(param.Description) < 1 {
		return ErrMissingParamDesc
	}

	if len(param.Type) < 1 || param.Type == "?" {
		return ErrMissingParamType
	}

	// optional type
	if param.Type[0] == '?' {
		param.Optional = true
		param.Type = param.Type[1:]
	}

	// find validator
	for _, validator := range validators {
		param.Validator = validator.Validator(param.Type, validators...)
		param.GoType = validator.GoType()
		if param.Validator != nil {
			break
		}
	}
	if param.Validator == nil {
		return ErrUnknownParamType
	}
	return nil
}
