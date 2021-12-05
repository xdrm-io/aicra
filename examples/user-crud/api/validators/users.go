package validators

import (
	"reflect"

	"github.com/xdrm-io/aicra/examples/user-crud/storage"
	"github.com/xdrm-io/aicra/validator"
)

// Users validator
type Users struct{}

// GoType returns the interface{} type
func (Users) GoType() reflect.Type {
	return reflect.TypeOf([]storage.User{})
}

// Validator that considers any value valid
func (Users) Validator(typename string, avail ...validator.Type) validator.ValidateFunc {
	if typename != "[]user" {
		return nil
	}

	// no need to validate ; only used as output
	return func(value interface{}) (interface{}, bool) {
		return value, true
	}
}
