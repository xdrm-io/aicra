package multipart

// cerr allows you to create constant "const" error with type boxing.
type cerr string

// Error implements the error builtin interface.
func (err cerr) Error() string {
	return string(err)
}

// ErrMissingDataName is set when a multipart variable/file has no name="..."
const ErrMissingDataName = cerr("data has no name")

// ErrDataNameConflict is set when a multipart variable/file name is already used
const ErrDataNameConflict = cerr("data name conflict")

// ErrNoHeader is set when a multipart variable/file has no (valid) header
const ErrNoHeader = cerr("data has no header")

// Component represents a multipart variable/file
type Component struct {
	// Content Type (raw for variables ; exported from files)
	ContentType string

	// data headers
	Headers map[string]string

	// actual data
	Data []byte
}
