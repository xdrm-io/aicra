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

func (param *Parameter) validate(datatypes ...validator.Type) error {
	if len(param.Description) < 1 {
		return errMissingParamDesc
	}

	if len(param.Type) < 1 || param.Type == "?" {
		return errMissingParamType
	}

	// optional type
	if param.Type[0] == '?' {
		param.Optional = true
		param.Type = param.Type[1:]
	}

	// find validator
	for _, dtype := range datatypes {
		param.Validator = dtype.Validator(param.Type, datatypes...)
		param.GoType = dtype.GoType()
		if param.Validator != nil {
			break
		}
	}
	if param.Validator == nil {
		return errUnknownDataType
	}
	return nil
}
