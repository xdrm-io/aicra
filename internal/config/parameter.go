package config

import (
	"reflect"

	"git.xdrm.io/go/aicra/datatype"
)

// Parameter represents a parameter definition (from api.json)
type Parameter struct {
	Description string `json:"info"`
	Type        string `json:"type"`
	Rename      string `json:"name,omitempty"`
	// ExtractType is the type of data the datatype returns
	ExtractType reflect.Type
	// Optional is set to true when the type is prefixed with '?'
	Optional bool

	// Validator is inferred from @Type
	Validator datatype.Validator
}

func (param *Parameter) validate(datatypes ...datatype.T) error {
	// missing description
	if len(param.Description) < 1 {
		return ErrMissingParamDesc
	}

	// invalid type
	if len(param.Type) < 1 || param.Type == "?" {
		return ErrMissingParamType
	}

	// optional type transform
	if param.Type[0] == '?' {
		param.Optional = true
		param.Type = param.Type[1:]
	}

	// assign the datatype
	for _, dtype := range datatypes {
		param.Validator = dtype.Build(param.Type, datatypes...)
		param.ExtractType = dtype.Type()
		if param.Validator != nil {
			break
		}
	}
	if param.Validator == nil {
		return ErrUnknownDataType
	}

	return nil
}
