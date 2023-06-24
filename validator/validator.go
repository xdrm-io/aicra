package validator

// ExtractFunc returns whether a given value fulfills the datatype and casts
// the value into a go type.
//
// for example, if a validator checks for upper case strings, whether the value
// is a []byte, a string or a []rune, if the value matches is all upper-case, it
// will be cast into a go type, say, string.
type ExtractFunc[T any] func(value any) (cast T, valid bool)

// Validator defines an available innput parameter "type" for the aicra
// configuration
//
// Every validator maps to a specific generic go type in order to generate the
// handler signature from the aicra configuration
//
// A Validator returns a custom extractor when the typename matches
type Validator[T any] interface {
	// Validate function when the typename matches. It must return nil when the
	// typename does not match
	//
	// The `typename` argument has to match types used in your aicra configuration
	// in parameter definitions ("in") and in the "type" json field.
	//
	// basic example:
	// - `Int.Validate("string")` should return nil
	// - `Int.Validate("int")` should return its ExtractFunc
	//
	// The `typename` is not returned by a simple method i.e. `TypeName() string`
	// because it allows for validation relative to the typename, for instance:
	// - `Varchar.Validate("varchar")` valides any string
	// - `Varchar.Validate("varchar(2)")` validates any string of 2
	//   characters
	// - `Varchar.Validate("varchar(1,3)")` validates any string
	//   with a length between 1 and 3
	//
	// The `avail` argument represents all other available Types. It allows a
	// Type to use other available Types internally.
	//
	// recursive example: slices
	// - `Slice.Validate("[]int", avail...)` validates a slice containing
	//   values that are valid to the `int` typename
	// - `Slice.Validate("[]varchar", avail...)` validates a slice containing
	//   values that are valid to the `varchar` type
	//
	// and so on.. this works for maps, structs, etc
	Validate(typename string) ExtractFunc[T]
}
