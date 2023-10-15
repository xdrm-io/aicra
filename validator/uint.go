package validator

import (
	"strconv"
)

// Uint considers valid:
// * uint
// * int (since it does not overflow)
// * float64 (since it does not overflow)
// * strings containing json-compatible integers
// * []byte containing json-compatible integers
type Uint struct{}

// Validate implements Validator for uint values
func (Uint) Validate(params []string) ExtractFunc[uint] {
	if len(params) != 0 {
		return nil
	}

	return func(value interface{}) (uint, bool) {
		switch cast := value.(type) {

		case int:
			return castNumber[int, uint](cast)
		case int8:
			return castNumber[int8, uint](cast)
		case int16:
			return castNumber[int16, uint](cast)
		case int32:
			return castNumber[int32, uint](cast)
		case int64:
			return castNumber[int64, uint](cast)
		case uint:
			return castNumber[uint, uint](cast)
		case uint8:
			return castNumber[uint8, uint](cast)
		case uint16:
			return castNumber[uint16, uint](cast)
		case uint32:
			return castNumber[uint32, uint](cast)
		case uint64:
			return castNumber[uint64, uint](cast)
		case float32:
			return castNumber[float32, uint](cast)
		case float64:
			return castNumber[float64, uint](cast)

			// serialized string -> try to convert to float
		case string:
			num, err := strconv.ParseUint(cast, 10, 64)
			return uint(num), err == nil

		case []byte:
			num, err := strconv.ParseUint(string(cast), 10, 64)
			return uint(num), err == nil

			// unknown type
		default:
			return 0, false
		}
	}
}
