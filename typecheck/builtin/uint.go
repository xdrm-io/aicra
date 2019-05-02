package builtin

import (
	"git.xdrm.io/go/aicra/typecheck"
)

// Uint checks if a value is an uint
type Uint struct{}

// NewUint returns a bare number type checker
func NewUint() *Uint {
	return &Uint{}
}

// Checker returns the checker function
func (Uint) Checker(typeName string) typecheck.CheckerFunc {
	// nothing if type not handled
	if typeName != "uint" {
		return nil
	}
	return func(value interface{}) bool {
		cast, isFloat := readFloat(value)

		if !isFloat {
			return false
		}

		return cast >= 0 && cast == float64(int(cast))
	}
}
