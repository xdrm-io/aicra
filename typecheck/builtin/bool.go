package builtin

import "git.xdrm.io/go/aicra/typecheck"

// Bool checks if a value is a boolean
type Bool struct{}

// NewBool returns a bare boolean type checker
func NewBool() *Bool {
	return &Bool{}
}

// Checker returns the checker function
func (Bool) Checker(typeName string) typecheck.Checker {
	// nothing if type not handled
	if typeName != "bool" {
		return nil
	}
	return func(value interface{}) bool {
		_, isBool := value.(bool)
		return isBool
	}
}
