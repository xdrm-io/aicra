package builtin

import (
	"git.xdrm.io/go/aicra/internal/checker"
)

// Any is a permissive type checker
type Any struct{}

// NewAny returns a bare any type checker
func NewAny() *Any {
	return &Any{}
}

// Checker returns the checker function
func (Any) Checker(typeName string) checker.Checker {
	// nothing if type not handled
	if typeName != "any" {
		return nil
	}
	return func(interface{}) bool {
		return true
	}
}
