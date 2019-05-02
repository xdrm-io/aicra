package typecheck

import "errors"

// ErrNoMatchingType when no available type checker matches the type
var ErrNoMatchingType = errors.New("no matching type")

// ErrDoesNotMatch when the value is invalid
var ErrDoesNotMatch = errors.New("does not match")

// CheckerFunc returns whether a given value fulfills a type
type CheckerFunc func(interface{}) bool

// Type represents a type checker
type Type interface {
	// given a type name, returns the checker function or NIL if the type is not handled here
	Checker(string) CheckerFunc
}
