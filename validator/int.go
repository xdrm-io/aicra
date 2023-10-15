package validator

import (
	"strconv"
)

// Int considers valid:
// * int
// * float64 (since it does not overflow)
// * uint (since it does not overflow)
// * strings containing json-compatible integers
// * []byte containing json-compatible integers
type Int struct{}

// Validate implements Validator
func (Int) Validate(params []string) ExtractFunc[int] {
	if len(params) != 0 {
		return nil
	}
	return func(value interface{}) (int, bool) {
		switch cast := value.(type) {

		case int:
			return castNumber[int, int](cast)
		case int8:
			return castNumber[int8, int](cast)
		case int16:
			return castNumber[int16, int](cast)
		case int32:
			return castNumber[int32, int](cast)
		case int64:
			return castNumber[int64, int](cast)
		case uint:
			return castNumber[uint, int](cast)
		case uint8:
			return castNumber[uint8, int](cast)
		case uint16:
			return castNumber[uint16, int](cast)
		case uint32:
			return castNumber[uint32, int](cast)
		case uint64:
			return castNumber[uint64, int](cast)
		case float32:
			return castNumber[float32, int](cast)
		case float64:
			return castNumber[float64, int](cast)

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
