package checker

// Matcher returns whether a type 'name' matches a type
type Matcher func(name string) bool

// Checker returns whether 'value' is valid to this Type
// note: it is a pointer because it can be formatted by the checker if matches
// to provide indulgent type check if needed
type Checker func(value interface{}) bool

// Type contains all necessary methods
// for a type provided by user/developer
type Type struct {
	Match func(string) bool
	Check func(interface{}) bool
}

// TypeRegistry represents a registry containing all available
// Type-s to be used by the framework according to the configuration
type TypeRegistry struct {
	Types []Type // registered Type-s
}
