package validator

import (
	"github.com/xdrm-io/aicra/examples/minimal/model"
	"github.com/xdrm-io/aicra/validator"
)

// Users is a nil validator used for output
type Users struct{}

// Validate implements aicra validator.Validator
func (Users) Validate(params []string) validator.ExtractFunc[[]model.User] {
	return func(value interface{}) ([]model.User, bool) {
		return []model.User{}, true
	}
}
