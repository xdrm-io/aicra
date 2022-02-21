package multipart

import (
	"bufio"
	"fmt"
	"io"
)

// DefaultContentType used when a component is NOT a file but literal data
const DefaultContentType = "raw"

// Reader is a multipart reader.
type Reader struct {
	// io.Reader used for reading multipart components reading.
	reader *bufio.Reader

	// boundary used to separate multipart components.
	boundary string

	// data will contain parsed components.
	Data map[string]*Component
}

// NewReader creates a new multipart reader for a reader and a boundary.
func NewReader(r io.Reader, boundary string) (*Reader, error) {
	reader := &Reader{
		reader:   nil,
		boundary: fmt.Sprintf("--%s", boundary),
		Data:     make(map[string]*Component),
	}

	// 1. Create reader
	dst, ok := r.(*bufio.Reader)
	if !ok {
		dst = bufio.NewReader(r)
	}
	reader.reader = dst

	// 2. "move" reader right after the first boundary
	var err error
	line := make([]byte, 0)

	for err == nil && string(line) != reader.boundary {
		line, _, err = dst.ReadLine()
	}
	if err != nil {
		return nil, err
	}

	// 3. return reader
	return reader, nil

}

// Parse parses the multipart components from the reader.
func (reader *Reader) Parse() error {

	// for each component (until boundary)
	for {

		comp := &Component{
			ContentType: "raw",
			Data:        make([]byte, 0),
			Headers:     make(map[string]string),
		}

		// 1. Read and parse data
		err := comp.read(reader.reader, reader.boundary)

		// 3. Dispatch error
		if err != nil && err != io.EOF {
			return err
		}

		name := comp.GetHeader("name")
		if len(name) < 1 {
			return errMissingDataName
		}

		if _, nameUsed := reader.Data[name]; nameUsed {
			return errDataNameConflict
		}

		reader.Data[name] = comp

		if err == io.EOF {
			return nil
		}

	}

}

// Get returns a multipart component by its name.
func (reader *Reader) Get(_key string) *Component {
	data, ok := reader.Data[_key]
	if !ok {
		return nil
	}
	return data
}
