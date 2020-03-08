package cerr

// Error allows you to create constant "const" error with type boxing.
type Error string

// Error implements the error builtin interface.
func (err Error) Error() string {
	return string(err)
}

// Wrap returns a new error which wraps a new error into itself.
func (err Error) Wrap(e error) *WrapError {
	return &WrapError{
		base: err,
		wrap: e,
	}
}

// WrapString returns a new error which wraps a new error created from a string.
func (err Error) WrapString(e string) *WrapError {
	return &WrapError{
		base: err,
		wrap: Error(e),
	}
}

// WrapError is way to wrap errors recursively.
type WrapError struct {
	base error
	wrap error
}

// Error implements the error builtin interface recursively.
func (err *WrapError) Error() string {
	return err.base.Error() + ": " + err.wrap.Error()
}
