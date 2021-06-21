package validator

import (
	"reflect"
)

// AnyType makes the "any" type available in the aicra configuration
// It considers valid any value
type AnyType struct{}

// GoType returns the interface{} type
func (AnyType) GoType() reflect.Type {
	return reflect.TypeOf(interface{}(nil))
}

// Validator that considers any value valid
func (AnyType) Validator(typename string, avail ...Type) ValidateFunc {
	if typename != "any" {
		return nil
	}
	return func(value interface{}) (interface{}, bool) {
		return value, true
	}
}
