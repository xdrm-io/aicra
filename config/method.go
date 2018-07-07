package config

import (
	"fmt"
	"git.xdrm.io/go/aicra/middleware"
)

// CheckScope returns whether a given scope matches the
// method configuration
//
// format is: [ [a,b], [c], [d,e] ]
// > level 1 is OR
// > level 2 is AND
func (m *Method) CheckScope(scope middleware.Scope) bool {
	fmt.Printf("Scope: %v\n", m.Permission)
	return false
}
