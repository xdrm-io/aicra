package aicra

// Err allows you to create constant "const" error with type boxing.
type Err string

func (e Err) Error() string {
	return string(e)
}

const (
	// ErrLateType - cannot add datatype after setting up the definition
	ErrLateType = Err("types cannot be added after Setup")

	// ErrNotSetup - not set up yet
	ErrNotSetup = Err("not set up")

	// ErrAlreadySetup - already set up
	ErrAlreadySetup = Err("already set up")

	// ErrUnknownService - no service matching this handler
	ErrUnknownService = Err("unknown service")

	// ErrAlreadyBound - handler already bound
	ErrAlreadyBound = Err("already bound")

	// ErrNilHandler - nil handler provided
	ErrNilHandler = Err("nil handler")

	// ErrMissingHandler - missing handler
	ErrMissingHandler = Err("missing handler")

	// ErrNilValidators - nil validators provided
	ErrNilValidators = Err("nil validators")

	// ErrNilResponder - nil responder provided
	ErrNilResponder = Err("nil responder")
)
