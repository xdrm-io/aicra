package validator

import (
	"math"
	"reflect"
	"strconv"
)

// IntType makes the "int" type available in the aicra configuration
// It considers valid:
// - int
// - float64 (since it does not overflow)
// - uint (since it does not overflow)
// - strings containing json-compatible integers
// - []byte containing json-compatible integers
type IntType struct{}

// GoType returns the `int` type
func (IntType) GoType() reflect.Type {
	return reflect.TypeOf(int(0))
}

// Validator for int values
func (IntType) Validator(typename string, avail ...Type) ValidateFunc {
	// nothing if type not handled
	if typename != "int" {
		return nil
	}

	return func(value interface{}) (interface{}, bool) {
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

			// serialized string -> try to convert to int
		case string:
			num, err := strconv.ParseInt(cast, 10, 64)
			return int(num), err == nil

			// serialized string -> try to convert to int
		case []byte:
			num, err := strconv.ParseInt(string(cast), 10, 64)
			return int(num), err == nil

			// unknown type
		default:
			return 0, false
		}
	}
}
