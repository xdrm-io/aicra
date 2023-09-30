package generated

import (
	model "github.com/xdrm-io/aicra/examples/minimal/model"
	custom "github.com/xdrm-io/aicra/examples/minimal/validator"
	builtin "github.com/xdrm-io/aicra/validator"
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

// Validators lists available validators for this API
var Validators = map[string]builtin.Validator[any]{
	"string": builtin.Wrap[string](new(builtin.String)),
	"users":  builtin.Wrap[[]model.User](new(custom.Users)),
	"uuid":   builtin.Wrap[string](new(custom.UUID)),
}
