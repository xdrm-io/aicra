package builtin

import (
	"encoding/json"

	"git.xdrm.io/go/aicra/datatype"
)

// FloatDataType is what its name tells
type FloatDataType struct{}

// Build returns the validator
func (FloatDataType) Build(typeName string) datatype.Validator {
	// nothing if type not handled
	if typeName != "float64" && typeName != "float" {
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
