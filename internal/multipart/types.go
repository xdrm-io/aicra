package multipart

import (
	"bufio"
)

// Reader represents a multipart reader
type Reader struct {
	// reader used for http.Request.Body reading
	reader *bufio.Reader

	// boundary used to separate multipart components
	boundary string

	// result will be inside this field
	Components map[string]*Component
}

// Component represents a multipart component
type Component struct {
	// whether this component is a file
	// if not, it is a simple variable data
	File bool

	// actual data
	Data []string
}
