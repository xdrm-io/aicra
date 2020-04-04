package aicra

// cerr allows you to create constant "const" error with type boxing.
type cerr string

func (err cerr) Error() string {
	return string(err)
}

// errLateType - cannot add datatype after setting up the definition
const errLateType = cerr("types cannot be added after Setup")

// errNotSetup - not set up yet
const errNotSetup = cerr("not set up")

// errAlreadySetup - already set up
const errAlreadySetup = cerr("already set up")

// errUnknownService - no service matching this handler
const errUnknownService = cerr("unknown service")

// errMissingHandler - missing handler
const errMissingHandler = cerr("missing handler")
