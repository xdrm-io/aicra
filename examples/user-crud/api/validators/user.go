package validators

import (
	"reflect"

	"github.com/xdrm-io/aicra/examples/user-crud/storage"
	"github.com/xdrm-io/aicra/validator"
)

// User validator
type User struct{}

// GoType returns the interface{} type
func (User) GoType() reflect.Type {
	return reflect.TypeOf(storage.User{})
}

// Validator that considers any value valid
func (User) Validator(typename string, avail ...validator.Type) validator.ValidateFunc {
	if typename != "user" {
		return nil
	}

	// no need to validate ; only used as output
	return func(value interface{}) (interface{}, bool) {
		return value, true
	}
}
