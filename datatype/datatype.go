package datatype

import (
	"reflect"
)

// Validator returns whether a given value fulfills the datatype
// and casts the value into a common go type.
//
// for example, if a validator checks for upper case strings,
// whether the value is a []byte, a string or a []rune, if the
// value matches the validator's checks, it will be cast it into
// a common go type, say, string.
type Validator func(value interface{}) (cast interface{}, valid bool)

// T represents a datatype. The Build function returns a Validator if
// it manages types with the name `typeDefinition` (from the configuration field "type"); else it or returns NIL if the type
// definition does not match this datatype; the registry is passed to allow recursive datatypes (e.g. slices, structs, etc)
// The datatype's validator (when input is valid) must return a cast's go type matching the `Type() reflect.Type`
type T interface {
	Type() reflect.Type
	Build(typeDefinition string, registry ...T) Validator
}
