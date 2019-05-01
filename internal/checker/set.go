package checker

// Set of type checkers
type Set struct {
	types []Type
}

// New returns a new set of type checkers
func New() *Set {
	return &Set{types: make([]Type, 0)}
}

// Add adds a new type checker
func (s *Set) Add(typeChecker Type) {
	s.types = append(s.types, typeChecker)
}

// Run finds a type checker from the registry matching the type `typeName`
// and uses this checker to check the `value`. If no type checker matches
// the `type`, error is returned by default.
func (s *Set) Run(typeName string, value interface{}) error {

	// find matching type (take first)
	for _, typeChecker := range s.types {
		if typeChecker == nil {
			continue
		}

		// found
		checkerFunc := typeChecker.Checker(typeName)
		if checkerFunc == nil {
			continue
		}

		// check value
		if checkerFunc(value) {
			return nil
		} else {
			return ErrDoesNotMatch
		}
	}

	return ErrNoMatchingType
}
