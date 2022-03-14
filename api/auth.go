package api

// Auth can be used by http middleware to
// 1) consult required roles in @Auth.Required
// 2) update active roles in @Auth.Active
type Auth struct {
	// required roles for this request
	// - the first dimension of the array reads as a OR
	// - the second dimension reads as a AND
	//
	// Example:
	// [ [A, B], [C, D] ] reads: roles (A and B) or (C and D) are required
	//
	// Warning: must not be mutated
	Required [][]string

	// active roles to be updated by authentication
	// procedures (e.g. jwt)
	Active []string
}

// Granted returns whether the authorization is granted
// i.e. Auth.Active fulfills Auth.Required
func (a *Auth) Granted() bool {
	var nothingRequired = true

	// first dimension: OR ; at least one is valid
	for _, required := range a.Required {
		// empty list
		if len(required) < 1 {
			continue
		}

		nothingRequired = false

		// second dimension: AND ; all required must be fulfilled
		if a.fulfills(required) {
			return true
		}
	}

	return nothingRequired
}

// returns whether Auth.Active fulfills (contains) all @required roles
func (a *Auth) fulfills(required []string) bool {
	if a.Active == nil {
		return false
	}
	for _, requiredRole := range required {
		var found = false
		for _, activeRole := range a.Active {
			if activeRole == requiredRole {
				found = true
				break
			}
		}
		// missing role -> fail
		if !found {
			return false
		}
	}
	// all @required are fulfilled
	return true
}
