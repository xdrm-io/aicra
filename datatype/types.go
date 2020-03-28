package datatype

import "reflect"

// Validator returns whether a given value fulfills a datatype
// and casts the value into a compatible type
type Validator func(value interface{}) (cast interface{}, valid bool)

// T builds a T from the type definition (from the configuration field "type") and returns NIL if the type
// definition does not match this T ; the registry is passed for recursive datatypes (e.g. slices, structs, etc)
// to be able to access other datatypes
type T interface {
	Type() reflect.Type
	Build(typeDefinition string, registry ...T) Validator
}
