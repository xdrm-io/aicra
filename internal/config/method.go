package config

import (
	"fmt"
	"strings"

	"git.xdrm.io/go/aicra/middleware"
)

// checkAndFormat checks for errors and missing fields and sets default values for optional fields.
func (methodDef *Method) checkAndFormat(servicePath string, httpMethod string) error {

	// 1. fail on missing description
	if len(methodDef.Description) < 1 {
		return fmt.Errorf("missing %s.%s description", servicePath, httpMethod)
	}

	// 2. stop if no parameter
	if methodDef.Parameters == nil || len(methodDef.Parameters) < 1 {
		methodDef.Parameters = make(map[string]*Parameter, 0)
		return nil
	}

	// 3. for each parameter
	for pName, pData := range methodDef.Parameters {

		// check name
		if strings.Trim(pName, "_") != pName {
			return fmt.Errorf("invalid name '%s' must not begin/end with '_'", pName)
		}

		if len(pData.Rename) < 1 {
			pData.Rename = pName
		}

		// 5. Check for name/rename conflict
		for paramName, param := range methodDef.Parameters {

			// ignore self
			if pName == paramName {
				continue
			}

			// 1. Same rename field
			if pData.Rename == param.Rename {
				return fmt.Errorf("rename conflict for %s.%s parameter '%s'", servicePath, httpMethod, pData.Rename)
			}

			// 2. Not-renamed field matches a renamed field
			if pName == param.Rename {
				return fmt.Errorf("name conflict for %s.%s parameter '%s'", servicePath, httpMethod, pName)
			}

			// 3. Renamed field matches name
			if pData.Rename == paramName {
				return fmt.Errorf("name conflict for %s.%s parameter '%s'", servicePath, httpMethod, pName)
			}

		}

		// 6. Manage invalid type
		if len(pData.Type) < 1 {
			return fmt.Errorf("invalid type for %s.%s parameter '%s'", servicePath, httpMethod, pName)
		}

		// 7. Fail on missing description
		if len(pData.Description) < 1 {
			return fmt.Errorf("missing description for %s.%s parameter '%s'", servicePath, httpMethod, pName)
		}

		// 8. Fail on missing type
		if len(pData.Type) < 1 {
			return fmt.Errorf("missing type for %s.%s parameter '%s'", servicePath, httpMethod, pName)
		}

		// 9. Set optional + type
		if pData.Type[0] == '?' {
			pData.Optional = true
			pData.Type = pData.Type[1:]
		}

	}

	return nil
}

// CheckScope returns whether a given scope matches the method configuration.
// The scope format is: `[ [a,b], [c], [d,e] ]` where the first level is a bitwise `OR` and the second a bitwise `AND`
func (methodDef *Method) CheckScope(scope middleware.Scope) bool {

	for _, OR := range methodDef.Permission {
		granted := true

		for _, AND := range OR {
			if !scopeHasPermission(AND, scope) {
				granted = false
				break
			}
		}

		// if one is valid -> grant
		if granted {
			return true
		}
	}

	return false
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
