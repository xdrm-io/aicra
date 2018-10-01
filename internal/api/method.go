package api

import (
	"git.xdrm.io/go/aicra/middleware"
)

// CheckScope returns whether a given scope matches the
// method configuration
//
// format is: [ [a,b], [c], [d,e] ]
// > level 1 is OR
// > level 2 is AND
func (m *Method) CheckScope(scope middleware.Scope) bool {

	for _, OR := range m.Permission {

		granted := true

		for _, AND := range OR {

			if !isPermInScope(AND, scope) {
				granted = false
				break
			}

		}

		// if one is valid -> grant
		if granted {
			return true
		}

	}

	return false
}

// Returns whether @perm is present in @scope
func isPermInScope(perm string, scope []string) bool {
	for _, s := range scope {
		if perm == s {
			return true
		}
	}
	return false
}
