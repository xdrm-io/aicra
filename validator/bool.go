package validator

import (
	"reflect"
)

// BoolType makes the "bool" type available in the aicra configuration
// It considers valid:
// - booleans
// - strings containing "true" or "false"
// - []byte containing "true" or "false"
type BoolType struct{}

// GoType returns the `bool` type
func (BoolType) GoType() reflect.Type {
	return reflect.TypeOf(true)
}

// Validator for bool values
func (BoolType) Validator(typename string, avail ...Type) ValidateFunc {
	if typename != "bool" {
		return nil
	}

	return func(value interface{}) (interface{}, bool) {
		switch cast := value.(type) {
		case bool:
			return cast, true

		case string:
			return cast == "true", cast == "true" || cast == "false"

		case []byte:
			strVal := string(cast)
			return strVal == "true", strVal == "true" || strVal == "false"

		default:
			return false, false
		}
	}
}
