package config

import "git.xdrm.io/go/aicra/config/datatype"

func (param *Parameter) checkAndFormat() error {

	// missing description
	if len(param.Description) < 1 {
		return ErrMissingParamDesc
	}

	// invalid type
	if len(param.Type) < 1 || param.Type == "?" {
		return ErrMissingParamType
	}

	// set optional + type
	if param.Type[0] == '?' {
		param.Optional = true
		param.Type = param.Type[1:]
	}

	return nil
}

// assigns the first matching data type from the type definition
func (param *Parameter) assignDataType(types []datatype.DataType) bool {
	for _, dtype := range types {
		param.Validator = dtype.Build(param.Type)
		if param.Validator != nil {
			return true
		}
	}
	return false
}
