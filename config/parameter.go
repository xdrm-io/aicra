package config

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
