package validator

import (
	"encoding/hex"

	"github.com/xdrm-io/aicra/validator"
)

// UUID valides unique identifiers
type UUID struct{}

// Validate implements aicra validator.Validator
func (UUID) Validate(params []string) validator.ExtractFunc[string] {
	return func(value interface{}) (string, bool) {
		str, isStr := value.(string)
		if bytes, ok := value.([]byte); ok {
			str = string(bytes)
			isStr = true
		}
		if !isStr {
			return "", false
		}

		// only accept 16 hex characters
		if len(str) != 16 {
			return "", false
		}
		if _, err := hex.DecodeString(str); err != nil {
			return "", false
		}
		return str, true
	}
}
