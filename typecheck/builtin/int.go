package builtin

import (
	"encoding/json"
	"math"

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
		_, isInt := readInt(value)

		return isInt
	}
}

// readInt tries to read a serialized int and returns whether it succeeded.
func readInt(value interface{}) (int, bool) {
	switch cast := value.(type) {

	case int:
		return cast, true

	case uint:
		overflows := cast > math.MaxInt64
		return int(cast), !overflows

	case float64:
		intVal := int(cast)
		overflows := cast < float64(math.MinInt64) || cast > float64(math.MaxInt64)
		return intVal, cast == float64(intVal) && !overflows

		// serialized string -> try to convert to float
	case string:
		num := json.Number(cast)
		intVal, err := num.Int64()
		return int(intVal), err == nil
		// serialized string -> try to convert to float

	case []byte:
		num := json.Number(cast)
		intVal, err := num.Int64()
		return int(intVal), err == nil

		// unknown type
	default:
		return 0, false

	}
}
