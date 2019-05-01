package builtin

import "git.xdrm.io/go/aicra/internal/checker"

// Float64 checks if a value is a float64
type Float64 struct{}

// NewFloat64 returns a bare number type checker
func NewFloat64() *Float64 {
	return &Float64{}
}

// Checker returns the checker function
func (Float64) Checker(typeName string) checker.Checker {
	// nothing if type not handled
	if typeName != "float64" && typeName != "float" {
		return nil
	}
	return func(value interface{}) bool {
		_, isFloat64 := value.(bool)
		return isFloat64
	}
}
