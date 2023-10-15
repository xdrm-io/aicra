package validator

// Bool considers valid
// * booleans
// * strings or []byte strictly equal to "true" or "false"
type Bool struct{}

// Validate implements Validator
func (Bool) Validate(params []string) ExtractFunc[bool] {
	if len(params) != 0 {
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
