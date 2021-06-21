package validator

import (
	"encoding/json"
	"math"
	"reflect"
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
			num := json.Number(cast)
			floatVal, err := num.Float64()
			if err != nil {
				return 0, false
			}
			overflows := floatVal < 0 || floatVal > math.MaxUint64
			return uint(floatVal), !overflows

		case []byte:
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
}
