package validator

import (
	"encoding/json"
	"reflect"
)

// FloatType makes the "float" (or "float64") type available in the aicra configuration
// It considers valid:
// - float64
// - int (since it does not overflow)
// - uint (since it does not overflow)
// - strings containing json-compatible floats
// - []byte containing json-compatible floats
type FloatType struct{}

// GoType returns the `float64` type
func (FloatType) GoType() reflect.Type {
	return reflect.TypeOf(float64(0))
}

// Validator for float64 values
func (FloatType) Validator(typename string, avail ...Type) ValidateFunc {
	if typename != "float64" && typename != "float" {
		return nil
	}
	return func(value interface{}) (interface{}, bool) {
		switch cast := value.(type) {

		case int:
			return float64(cast), true

		case uint:
			return float64(cast), true

		case float64:
			return cast, true

			// serialized string -> try to convert to float
		case []byte:
			num := json.Number(cast)
			floatVal, err := num.Float64()
			return floatVal, err == nil

		case string:
			num := json.Number(cast)
			floatVal, err := num.Float64()
			return floatVal, err == nil

			// unknown type
		default:
			return 0, false

		}
	}
}
