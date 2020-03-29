package aicra

// cerr allows you to create constant "const" error with type boxing.
type cerr string

// Error implements the error builtin interface.
func (err cerr) Error() string {
	return string(err)
}

// ErrNoServiceForHandler - no service matching this handler
const ErrNoServiceForHandler = cerr("no service found for this handler")

// ErrNoHandlerForService - no handler matching this service
const ErrNoHandlerForService = cerr("no handler found for this service")
