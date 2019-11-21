package config

import (
	"strings"
)

// checkAndFormat checks for errors and missing fields and sets default values for optional fields.
func (methodDef *Method) checkAndFormat(servicePath string, httpMethod string) error {

	// 1. fail on missing description
	if len(methodDef.Description) < 1 {
		return ErrMissingMethodDesc.WrapString(httpMethod + " " + servicePath)
	}

	// 2. stop if no parameter
	if methodDef.Parameters == nil || len(methodDef.Parameters) < 1 {
		methodDef.Parameters = make(map[string]*Parameter, 0)
		return nil
	}

	// 3. for each parameter
	for pName, pData := range methodDef.Parameters {

		// 3.1. check name
		if strings.Trim(pName, "_") != pName {
			return ErrIllegalParamName.WrapString(httpMethod + " " + servicePath + " {" + pName + "}")
		}

		if len(pData.Rename) < 1 {
			pData.Rename = pName
		}

		// 3.2. Check for name/rename conflict
		for paramName, param := range methodDef.Parameters {

			// ignore self
			if pName == paramName {
				continue
			}

			// 3.2.1. Same rename field
			if pData.Rename == param.Rename {
				return ErrParamNameConflict.WrapString(httpMethod + " " + servicePath + " {" + pName + "}")
			}

			// 3.2.2. Not-renamed field matches a renamed field
			if pName == param.Rename {
				return ErrParamNameConflict.WrapString(httpMethod + " " + servicePath + " {" + pName + "}")
			}

			// 3.2.3. Renamed field matches name
			if pData.Rename == paramName {
				return ErrParamNameConflict.WrapString(httpMethod + " " + servicePath + " {" + pName + "}")
			}

		}

		// 3.3. Fail on missing description
		if len(pData.Description) < 1 {
			return ErrMissingParamDesc.WrapString(httpMethod + " " + servicePath + " {" + pName + "}")
		}

		// 3.4. Manage invalid type
		if len(pData.Type) < 1 {
			return ErrMissingParamType.WrapString(httpMethod + " " + servicePath + " {" + pName + "}")
		}

		// 3.5. Set optional + type
		if pData.Type[0] == '?' {
			pData.Optional = true
			pData.Type = pData.Type[1:]
		}

	}

	return nil
}

// scopeHasPermission returns whether the permission fulfills a given scope
func scopeHasPermission(permission string, scope []string) bool {
	for _, s := range scope {
		if permission == s {
			return true
		}
	}
	return false
}
