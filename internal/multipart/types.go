package multipart

import (
	"bufio"
)

// ConstError is a wrapper to set constant errors
type ConstError string

// Error implements error
func (err ConstError) Error() string {
	return string(err)
}

// ErrMissingDataName is set when a multipart variable/file has no name="..."
var ErrMissingDataName = ConstError("data has no name")

// ErrDataNameConflict is set when a multipart variable/file name is already used
var ErrDataNameConflict = ConstError("data name conflict")

// ErrNoHeader is set when a multipart variable/file has no (valid) header
var ErrNoHeader = ConstError("data has no header")

// Component represents a multipart variable/file
type Component struct {
	// Content Type (raw for variables ; exported from files)
	ContentType string

	// data headers
	Headers map[string]string

	// actual data
	Data []byte
}

// Reader represents a multipart reader
type Reader struct {
	// reader used for http.Request.Body reading
	reader *bufio.Reader

	// boundary used to separate multipart MultipartDatas
	boundary string

	// result will be inside this field
	Data map[string]*Component
}
