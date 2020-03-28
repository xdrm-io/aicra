package builtin

import (
	"reflect"

	"git.xdrm.io/go/aicra/datatype"
)

// AnyDataType is what its name tells
type AnyDataType struct{}

// Type returns the type of data
func (AnyDataType) Type() reflect.Type {
	return reflect.TypeOf(interface{}(nil))
}

// Build returns the validator
func (AnyDataType) Build(typeName string, registry ...datatype.T) datatype.Validator {
	// nothing if type not handled
	if typeName != "any" {
		return nil
	}
	return func(value interface{}) (interface{}, bool) {
		return value, true
	}
}
