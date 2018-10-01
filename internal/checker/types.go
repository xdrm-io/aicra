package checker

import (
	"git.xdrm.io/go/aicra/driver"
)

// Matcher returns whether a type 'name' matches a type
type Matcher func(name string) bool

// Checker returns whether 'value' is valid to this Type
// note: it is a pointer because it can be formatted by the checker if matches
// to provide indulgent type check if needed
type Checker func(value interface{}) bool

// Registry represents a registry containing all available
// Type-s to be used by the framework according to the configuration
type Registry map[string]driver.Checker
