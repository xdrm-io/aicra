package builtin

import (
	"strconv"

	"git.xdrm.io/go/aicra/typecheck"
)

// Float64 checks if a value is a float64
type Float64 struct{}

// NewFloat64 returns a bare number type checker
func NewFloat64() *Float64 {
	return &Float64{}
}

// Checker returns the checker function
func (Float64) Checker(typeName string) typecheck.CheckerFunc {
	// nothing if type not handled
	if typeName != "float64" && typeName != "float" {
		return nil
	}
	return func(value interface{}) bool {
		_, isFloat := readFloat(value)
		return isFloat
	}
}

// readFloat tries to read a serialized float and returns whether it succeeded.
func readFloat(value interface{}) (float64, bool) {
	switch cast := value.(type) {

	case int:
		return float64(cast), true

	case uint:
		return float64(cast), true

	case float64:
		return cast, true

		// serialized string -> try to convert to float
	case string:
		floatVal, err := strconv.ParseFloat(cast, 64)
		return floatVal, err == nil

		// unknown type
	default:
		return 0, false

	}
}
