package multipart

import "git.xdrm.io/go/aicra/internal/cerr"

// ErrMissingDataName is set when a multipart variable/file has no name="..."
const ErrMissingDataName = cerr.Error("data has no name")

// ErrDataNameConflict is set when a multipart variable/file name is already used
const ErrDataNameConflict = cerr.Error("data name conflict")

// ErrNoHeader is set when a multipart variable/file has no (valid) header
const ErrNoHeader = cerr.Error("data has no header")

// Component represents a multipart variable/file
type Component struct {
	// Content Type (raw for variables ; exported from files)
	ContentType string

	// data headers
	Headers map[string]string

	// actual data
	Data []byte
}
