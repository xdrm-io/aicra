package validator

// Any makes the "any" type available in the aicra configuration
// It considers valid any value
type Any struct{}

// Validate implements Validator
func (Any) Validate(typename string) ExtractFunc[any] {
	if typename != "any" {
		return nil
	}
	return func(value interface{}) (any, bool) {
		return value, true
	}
}
