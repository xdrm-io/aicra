package generated

import (
	builtin "github.com/xdrm-io/aicra/validator"
	custom "github.com/xdrm-io/aicra/examples/minimal/validator"
	model "github.com/xdrm-io/aicra/examples/minimal/model"
)

func getBuiltinStringValidator(params []string) builtin.ExtractFunc[string] {
	return new(builtin.String).Validate(params)
}
func getCustomUsersValidator(params []string) builtin.ExtractFunc[[]model.User] {
	return new(custom.Users).Validate(params)
}
func getCustomUUIDValidator(params []string) builtin.ExtractFunc[string] {
	return new(custom.UUID).Validate(params)
}
