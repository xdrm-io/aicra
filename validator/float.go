package validator

import (
	"strconv"
)

// Float makes the "float" (or "float64") type available in the aicra configuration
// It considers valid:
// - float64
// - int (since it does not overflow)
// - uint (since it does not overflow)
// - strings containing json-compatible floats
// - []byte containing json-compatible floats
type Float struct{}

// Validate implements Validator
func (Float) Validate(typename string) ExtractFunc[float64] {
	if typename != "float64" && typename != "float" {
		return nil
	}
	return func(value interface{}) (float64, bool) {
		switch cast := value.(type) {

		case int:
			return castNumber[int, float64](cast)
		case int8:
			return castNumber[int8, float64](cast)
		case int16:
			return castNumber[int16, float64](cast)
		case int32:
			return castNumber[int32, float64](cast)
		case int64:
			return castNumber[int64, float64](cast)
		case uint:
			return castNumber[uint, float64](cast)
		case uint8:
			return castNumber[uint8, float64](cast)
		case uint16:
			return castNumber[uint16, float64](cast)
		case uint32:
			return castNumber[uint32, float64](cast)
		case uint64:
			return castNumber[uint64, float64](cast)
		case float32:
			return castNumber[float32, float64](cast)
		case float64:
			return castNumber[float64, float64](cast)

			// serialized string -> try to convert to float
		case []byte:
			num, err := strconv.ParseFloat(string(cast), 64)
			return num, err == nil

		case string:
			num, err := strconv.ParseFloat(cast, 64)
			return num, err == nil

			// unknown type
		default:
			return 0, false

		}
	}
}
