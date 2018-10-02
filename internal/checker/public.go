package checker

import (
	"errors"
	"git.xdrm.io/go/aicra/driver"
)

var ErrNoMatchingType = errors.New("no matching type")
var ErrDoesNotMatch = errors.New("does not match")

// CreateRegistry creates an empty type registry
func CreateRegistry() Registry {
	return make(Registry)
}

// Add adds a new checker for a path
func (reg Registry) Add(_path string, _element driver.Checker) {
	reg[_path] = _element
}

// Run finds a type checker from the registry matching the type @typeName
// and uses this checker to check the @value. If no type checker matches
// the @typeName name, error is returned by default.
func (reg Registry) Run(typeName string, value interface{}) error {

	/* (1) Iterate to find matching type (take first) */
	for _, t := range reg {

		if t == nil {
			continue
		}

		// stop if found
		if t.Match(typeName) {

			// check value
			if t.Check(value) {
				return nil
			}

			return ErrDoesNotMatch

		}

	}

	return ErrNoMatchingType

}
