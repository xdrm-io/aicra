package config

import (
	"reflect"

	"github.com/xdrm-io/aicra/datatype"
)

// Parameter represents a parameter definition (from api.json)
type Parameter struct {
	Description string `json:"info"`
	Type        string `json:"type"`
	Rename      string `json:"name,omitempty"`
	Optional    bool
	// ExtractType is the type the Validator will cast into
	ExtractType reflect.Type
	// Validator is inferred from the "type" property
	Validator datatype.Validator
}

func (param *Parameter) validate(datatypes ...datatype.T) error {
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
		param.Validator = dtype.Build(param.Type, datatypes...)
		param.ExtractType = dtype.Type()
		if param.Validator != nil {
			break
		}
	}
	if param.Validator == nil {
		return errUnknownDataType
	}
	return nil
}
