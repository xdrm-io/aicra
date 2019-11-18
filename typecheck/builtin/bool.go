package builtin

import "git.xdrm.io/go/aicra/typecheck"

// Bool checks if a value is a boolean
type Bool struct{}

// NewBool returns a bare boolean type checker
func NewBool() *Bool {
	return &Bool{}
}

// Checker returns the checker function
func (Bool) Checker(typeName string) typecheck.CheckerFunc {
	// nothing if type not handled
	if typeName != "bool" {
		return nil
	}
	return func(value interface{}) bool {
		_, isBool := readBool(value)
		return isBool
	}
}

// readBool tries to read a serialized boolean and returns whether it succeeded.
func readBool(value interface{}) (bool, bool) {
	switch cast := value.(type) {
	case bool:
		return cast, true

	case string:
		strVal := string(cast)
		return strVal == "true", strVal == "true" || strVal == "false"

	case []byte:
		strVal := string(cast)
		return strVal == "true", strVal == "true" || strVal == "false"

	default:
		return false, false
	}

	return false, false
}
