package multipart

// cerr allows you to create constant "const" error with type boxing.
type cerr string

func (err cerr) Error() string {
	return string(err)
}

// errMissingDataName is set when a multipart variable/file has no name="..."
const errMissingDataName = cerr("data has no name")

// errDataNameConflict is set when a multipart variable/file name is already used
const errDataNameConflict = cerr("data name conflict")

// errNoHeader is set when a multipart variable/file has no (valid) header
const errNoHeader = cerr("data has no header")

// Component represents a multipart variable/file
type Component struct {
	// Content Type (raw for variables ; exported from files)
	ContentType string

	// data headers
	Headers map[string]string

	// actual data
	Data []byte
}
