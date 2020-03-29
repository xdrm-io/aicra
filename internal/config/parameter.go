package config

import (
	"git.xdrm.io/go/aicra/datatype"
)

// Validate implements the validator interface
func (param *Parameter) Validate(datatypes ...datatype.T) error {
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
