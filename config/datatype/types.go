package datatype

// Validator returns whether a given value fulfills a datatype
// and casts the value into a compatible type
type Validator func(value interface{}) (cast interface{}, valid bool)

// Builder builds a DataType from the type definition (from the
// configuration field "type") and returns NIL if the type
// definition does not match this DataType
type Builder interface {
	Build(typeDefinition string) Validator
}
