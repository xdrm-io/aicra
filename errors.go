package aicra

// cerr allows you to create constant "const" error with type boxing.
type cerr string

// Error implements the error builtin interface.
func (err cerr) Error() string {
	return string(err)
}

// ErrLateType - cannot add datatype after setting up the definition
const ErrLateType = cerr("types cannot be added after Setup")

// ErrNotSetup - not set up yet
const ErrNotSetup = cerr("not set up")

// ErrAlreadySetup - already set up
const ErrAlreadySetup = cerr("already set up")

// ErrUnknownService - no service matching this handler
const ErrUnknownService = cerr("unknown service")

// ErrMissingHandler - missing handler
const ErrMissingHandler = cerr("missing handler")
