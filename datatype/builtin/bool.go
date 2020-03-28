package builtin

import (
	"reflect"

	"git.xdrm.io/go/aicra/datatype"
)

// BoolDataType is what its name tells
type BoolDataType struct{}

// Kind returns the kind of data
func (BoolDataType) Kind() reflect.Kind {
	return reflect.Bool
}

// Build returns the validator
func (BoolDataType) Build(typeName string, registry ...datatype.T) datatype.Validator {
	// nothing if type not handled
	if typeName != "bool" {
		return nil
	}

	return func(value interface{}) (interface{}, bool) {
		switch cast := value.(type) {
		case bool:
			return cast, true

		case string:
			strVal := string(cast)
			return strVal == "true", strVal == "true" || strVal == "false"
		case []byte:
			strVal := string(cast)
			return strVal == "true", strVal == "true" || strVal == "false"

		default:
			return false, false
		}
	}
}
