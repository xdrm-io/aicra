package multipart

import (
	"bufio"
	"fmt"
	"io"
)

// NewReader craetes a new reader
func NewReader(_src io.Reader, _boundary string) (*Reader, error) {

	reader := &Reader{
		reader:   nil,
		boundary: fmt.Sprintf("--%s", _boundary),
		Data:     make(map[string]*Component),
	}

	// 1. Create reader
	dst, ok := _src.(*bufio.Reader)
	if !ok {
		dst = bufio.NewReader(_src)
	}
	reader.reader = dst

	// 2. Place reader after the first boundary
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

// Parse parses the multipart components from the request
func (reader *Reader) Parse() error {

	/* (1) For each component (until boundary) */
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
			return ErrMissingDataName
		}

		if _, nameUsed := reader.Data[name]; nameUsed {
			return ErrDataNameConflict
		}

		reader.Data[name] = comp

		if err == io.EOF {
			return nil
		}

	}

}

// Get returns a multipart data by name, nil if not found
func (reader *Reader) Get(_key string) *Component {
	data, ok := reader.Data[_key]
	if !ok {
		return nil
	}
	return data
}
