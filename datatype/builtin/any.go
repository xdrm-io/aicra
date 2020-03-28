package builtin

import (
	"reflect"

	"git.xdrm.io/go/aicra/datatype"
)

// AnyDataType is what its name tells
type AnyDataType struct{}

// Kind returns the kind of data
func (AnyDataType) Kind() reflect.Kind {
	return reflect.Interface
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
