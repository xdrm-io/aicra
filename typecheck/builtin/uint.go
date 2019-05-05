package builtin

import (
	"encoding/json"
	"math"

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
		_, isInt := readUint(value)

		return isInt
	}
}

// readUint tries to read a serialized uint and returns whether it succeeded.
func readUint(value interface{}) (uint, bool) {
	switch cast := value.(type) {

	case int:
		return uint(cast), cast >= 0

	case uint:
		return cast, true

	case float64:
		uintVal := uint(cast)
		overflows := cast < 0 || cast > math.MaxUint64
		return uintVal, cast == float64(uintVal) && !overflows

		// serialized string -> try to convert to float
	case string:
		num := json.Number(cast)
		floatVal, err := num.Float64()
		if err != nil {
			return 0, false
		}
		overflows := floatVal < 0 || floatVal > math.MaxUint64
		return uint(floatVal), !overflows

		// unknown type
	default:
		return 0, false

	}
}
