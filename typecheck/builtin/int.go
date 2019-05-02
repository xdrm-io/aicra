package builtin

import (
	"git.xdrm.io/go/aicra/typecheck"
)

// Int checks if a value is an int
type Int struct{}

// NewInt returns a bare number type checker
func NewInt() *Int {
	return &Int{}
}

// Checker returns the checker function
func (Int) Checker(typeName string) typecheck.CheckerFunc {
	// nothing if type not handled
	if typeName != "int" {
		return nil
	}
	return func(value interface{}) bool {
		cast, isFloat := readFloat(value)

		if !isFloat {
			return false
		}

		return cast == float64(int(cast))
	}
}
