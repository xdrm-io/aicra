package validator

// ExtractFunc returns whether a given value fulfills the datatype and casts
// the value into a go type.
//
// for example, if a validator checks for upper case strings, whether the value
// is a []byte, a string or a []rune, if the value matches is all upper-case, it
// will be cast into a go type, say, string.
type ExtractFunc[T any] func(value any) (cast T, valid bool)

// Validator defines an available input parameter "type" for the aicra
// configuration
//
// Every validator maps to a specific generic go type in order to generate the
// handler signature from the aicra configuration
//
// A Validator returns a custom extractor when the params are valid
type Validator[T any] interface {
	// Validate function. It must return nil when params are invalid.
	//
	// The `params` argument has to match what is handled by the validator.
	// Params are extracted from the configuration syntax in parameter
	// definition ("in") and in the "type" json field.
	//
	// The format is as follows:
	// - "typename" -> no param
	// - "typename(a)" -> 1 param "a"
	// - "typename(a,b)" -> 2 params "a" and "b"
	//
	Validate(params []string) ExtractFunc[T]
}

// Wrap any generic validator into a fixed-type any validator
func Wrap[T any](v Validator[T]) Validator[any] {
	return wrapper(func(params []string) ExtractFunc[any] {
		extractor := v.Validate(params)
		if extractor == nil {
			return nil
		}
		return func(value any) (any, bool) {
			cast, ok := extractor(value)
			return any(cast), ok
		}
	})
}

// wraps a generic validator into a any validators
type wrapper func([]string) ExtractFunc[any]

// Validate for wrapped validator
func (w wrapper) Validate(params []string) ExtractFunc[any] {
	return (func([]string) ExtractFunc[any])(w)(params)
}
