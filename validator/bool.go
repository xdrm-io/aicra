package validator

// Bool makes the "bool" type available in the aicra configuration
// It considers valid:
// - booleans
// - strings containing "true" or "false"
// - []byte containing "true" or "false"
type Bool struct{}

// Validate implements Validator
func (Bool) Validate(typename string) ExtractFunc[bool] {
	if typename != "bool" {
		return nil
	}
	return func(value interface{}) (bool, bool) {
		switch cast := value.(type) {
		case bool:
			return cast, true
		case string:
			return cast == "true", cast == "true" || cast == "false"
		case []byte:
			str := string(cast)
			return str == "true", str == "true" || str == "false"
		default:
			return false, false
		}
	}
}
