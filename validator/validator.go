package validator

import (
	"reflect"
)

// ValidateFunc returns whether a given value fulfills the datatype and casts
// the value into a go type.
//
// for example, if a validator checks for upper case strings, whether the value
// is a []byte, a string or a []rune, if the value matches is all upper-case, it
// will be cast into a go type, say, string.
type ValidateFunc func(value interface{}) (cast interface{}, valid bool)

// Type defines an available innput parameter "type" for the aicra configuration
//
// A Type maps to a go type in order to generate the handler signature from the
// aicra configuration
//
// A Type returns a custom validator when the typename matches
type Type interface {
	// Validator function when the typename matches. It must return nil when the
	// typename does not match
	//
	// The `typename` argument has to match types used in your aicra configuration
	// in parameter definitions ("in") and in the "type" json field.
	//
	// basic example:
	// - `IntType.Validator("string")` should return nil
	// - `IntType.Validator("int")` should return its ValidateFunc
	//
	// The `typename` is not returned by a simple method i.e. `TypeName() string`
	// because it allows for validation relative to the typename, for instance:
	// - `VarcharType.Validator("varchar")` valides any string
	// - `VarcharType.Validator("varchar(2)")` validates any string of 2
	//   characters
	// - `VarcharType.Validator("varchar(1,3)")` validates any string
	//   with a length between 1 and 3
	//
	// The `avail` argument represents all other available Types. It allows a
	// Type to use other available Types internally.
	//
	// recursive example: slices
	// - `SliceType.Validator("[]int", avail...)` validates a slice containing
	//   values that are valid to the `int` typename
	// - `SliceType.Validator("[]varchar", avail...)` validates a slice containing
	//   values that are valid to the `varchar` type
	//
	// and so on.. this works for maps, structs, etc
	Validator(typename string, avail ...Type) ValidateFunc

	// GoType must return the go type associated with the output type of ValidateFunc.
	// It is used to define handlers' signature from the configuration file.
	GoType() reflect.Type
}
