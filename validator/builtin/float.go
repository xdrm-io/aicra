package builtin

import (
	"encoding/json"
	"reflect"

	"github.com/xdrm-io/aicra/validator"
)

// FloatDataType is what its name tells
type FloatDataType struct{}

// GoType returns the type of data
func (FloatDataType) GoType() reflect.Type {
	return reflect.TypeOf(float64(0))
}

// Validator returns the validator
func (FloatDataType) Validator(typeName string, registry ...validator.Type) validator.ValidateFunc {
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
