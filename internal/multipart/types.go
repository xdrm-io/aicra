package multipart

// ConstError is a wrapper to set constant errors
type ConstError string

// Error implements error
func (err ConstError) Error() string {
	return string(err)
}

// ErrMissingDataName is set when a multipart variable/file has no name="..."
const ErrMissingDataName = ConstError("data has no name")

// ErrDataNameConflict is set when a multipart variable/file name is already used
const ErrDataNameConflict = ConstError("data name conflict")

// ErrNoHeader is set when a multipart variable/file has no (valid) header
const ErrNoHeader = ConstError("data has no header")

// Component represents a multipart variable/file
type Component struct {
	// Content Type (raw for variables ; exported from files)
	ContentType string

	// data headers
	Headers map[string]string

	// actual data
	Data []byte
}
