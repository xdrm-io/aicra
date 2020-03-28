package builtin

import (
	"encoding/json"
	"math"

	"git.xdrm.io/go/aicra/datatype"
)

// IntDataType is what its name tells
type IntDataType struct{}

// Build returns the validator
func (IntDataType) Build(typeName string, registry ...datatype.T) datatype.Validator {
	// nothing if type not handled
	if typeName != "int" {
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
}
