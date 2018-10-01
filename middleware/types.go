package middleware

import (
	"git.xdrm.io/go/aicra/driver"
)

// Scope represents a list of scope processed by middlewares
// and used by the router to block/allow some uris
// it is also passed to controllers
//
// DISCLAIMER: it is used to help developers but for compatibility
//             purposes, the type is always used as its definition ([]string)
type Scope []string

// Registry represents a registry containing all registered
// middlewares to be processed before routing any request
// The map is <name> => <middleware>
type Registry map[string]driver.Middleware
