package builtin

import (
	"reflect"

	"github.com/xdrm-io/aicra/validator"
)

// BoolDataType is what its name tells
type BoolDataType struct{}

// GoType returns the type of data
func (BoolDataType) GoType() reflect.Type {
	return reflect.TypeOf(true)
}

// Validator returns the validator
func (BoolDataType) Validator(typeName string, registry ...validator.Type) validator.ValidateFunc {
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
