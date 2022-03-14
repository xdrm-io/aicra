package validator

import (
	"math"
	"reflect"
	"strconv"
)

// UintType makes the "uint" type available in the aicra configuration
// It considers valid:
// - uint
// - int (since it does not overflow)
// - float64 (since it does not overflow)
// - strings containing json-compatible integers
// - []byte containing json-compatible integers
type UintType struct{}

// GoType returns the `uint` type
func (UintType) GoType() reflect.Type {
	return reflect.TypeOf(uint(0))
}

// Validator for uint values
func (UintType) Validator(other string, avail ...Type) ValidateFunc {
	if other != "uint" {
		return nil
	}

	return func(value interface{}) (interface{}, bool) {
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
			num, err := strconv.ParseUint(cast, 10, 64)
			return int(num), err == nil

		case []byte:
			num, err := strconv.ParseUint(string(cast), 10, 64)
			return int(num), err == nil

			// unknown type
		default:
			return 0, false
		}
	}
}
