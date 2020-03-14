package builtin

import (
	"encoding/json"
	"math"

	"git.xdrm.io/go/aicra/config/datatype"
)

// UintDataType is what its name tells
type UintDataType struct{}

// Build returns the validator
func (UintDataType) Build(typeName string) datatype.Validator {
	// nothing if type not handled
	if typeName != "uint" {
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
