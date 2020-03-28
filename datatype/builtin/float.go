package builtin

import (
	"encoding/json"
	"reflect"

	"git.xdrm.io/go/aicra/datatype"
)

// FloatDataType is what its name tells
type FloatDataType struct{}

// Kind returns the kind of data
func (FloatDataType) Kind() reflect.Kind {
	return reflect.Float64
}

// Build returns the validator
func (FloatDataType) Build(typeName string, registry ...datatype.T) datatype.Validator {
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
