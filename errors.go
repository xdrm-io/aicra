package aicra

// cerr allows you to create constant "const" error with type boxing.
type cerr string

func (err cerr) Error() string {
	return string(err)
}

const (
	// errLateType - cannot add datatype after setting up the definition
	errLateType = cerr("types cannot be added after Setup")

	// errNotSetup - not set up yet
	errNotSetup = cerr("not set up")

	// errAlreadySetup - already set up
	errAlreadySetup = cerr("already set up")

	// errUnknownService - no service matching this handler
	errUnknownService = cerr("unknown service")

	// errUncallableService - nil handler provided
	errNilHandler = cerr("nil handler")

	// errMissingHandler - missing handler
	errMissingHandler = cerr("missing handler")

	// errNilResponder - nil responder provided
	errNilResponder = cerr("nil responder")
)
