package multipart

// Error allows you to create constant "const" error with type boxing.
type Error string

// Error implements the error builtin interface.
func (err Error) Error() string {
	return string(err)
}

// ErrMissingDataName is set when a multipart variable/file has no name="..."
const ErrMissingDataName = Error("data has no name")

// ErrDataNameConflict is set when a multipart variable/file name is already used
const ErrDataNameConflict = Error("data name conflict")

// ErrNoHeader is set when a multipart variable/file has no (valid) header
const ErrNoHeader = Error("data has no header")

// Component represents a multipart variable/file
type Component struct {
	// Content Type (raw for variables ; exported from files)
	ContentType string

	// data headers
	Headers map[string]string

	// actual data
	Data []byte
}
