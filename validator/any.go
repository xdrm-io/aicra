package validator

// Any considers valid any value
type Any struct{}

// Validate implements Validator
func (Any) Validate(params []string) ExtractFunc[any] {
	if len(params) != 0 {
		return nil
	}
	return func(value interface{}) (any, bool) {
		return value, true
	}
}
