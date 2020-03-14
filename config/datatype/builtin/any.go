package builtin

import "git.xdrm.io/go/aicra/config/datatype"

// AnyDataType is what its name tells
type AnyDataType struct{}

// Build returns the validator
func (AnyDataType) Build(typeName string) datatype.Validator {
	// nothing if type not handled
	if typeName != "any" {
		return nil
	}
	return func(value interface{}) (interface{}, bool) {
		return value, true
	}
}
