package multipart

import (
	"bufio"
)

type MultipartReader struct {
	// reader used for http.Request.Body reading
	reader *bufio.Reader

	// boundary used to separate multipart components
	boundary string

	// result will be inside this field
	Components map[string]*MultipartComponent
}

// Represents a multipart component
type MultipartComponent struct {
	// whether this component is a file
	// if not, it is a simple variable data
	File bool

	// actual data
	Data []string
}
