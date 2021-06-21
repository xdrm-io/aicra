package builtin

import (
	"reflect"

	"github.com/xdrm-io/aicra/validator"
)

// AnyDataType is what its name tells
type AnyDataType struct{}

// GoType returns the type of data
func (AnyDataType) GoType() reflect.Type {
	return reflect.TypeOf(interface{}(nil))
}

// Validator returns the validator
func (AnyDataType) Validator(typeName string, registry ...validator.Type) validator.ValidateFunc {
	// nothing if type not handled
	if typeName != "any" {
		return nil
	}
	return func(value interface{}) (interface{}, bool) {
		return value, true
	}
}
