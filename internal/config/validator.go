package config

import "github.com/xdrm-io/aicra/validator"

// Validators matches generated validators that implement validator.Validator.
// Generics do not allow using generic implementations without knowing the
// concrete type ; so we use "any" at runtime. It is only used for configuration
// validation.
type Validators map[string]validator.Validator[any]

// Validator defines a validator GO symbol and associated GO type
type Validator struct {
	Validator string `json:"use"`
	Type      string `json:"as"`
}
