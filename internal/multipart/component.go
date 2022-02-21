package multipart

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// parseHeaders parses a component headers.
func (comp *Component) parseHeaders(raw []byte) error {
	// 1. Extract lines
	lines := strings.Split(string(raw), "\n")
	if len(lines) < 2 {
		return errNoHeader
	}

	// 2. trim each line + remove 'Content-Disposition' prefix
	header := strings.Trim(lines[0], " \t\r")

	if !strings.HasPrefix(header, "Content-Disposition: form-data;") {
		return errNoHeader
	}
	header = strings.Trim(header[len("Content-Disposition: form-data;"):], " \t\r")

	if len(header) < 1 {
		return errNoHeader
	}

	// 3. Extract each key-value pair
	pairs := strings.Split(header, "; ")

	// 4. extract each pair
	for _, p := range pairs {
		pair := strings.Split(p, "=")

		// ignore invalid pairs
		if len(pair) != 2 || len(pair[1]) < 1 {
			continue
		}

		key := strings.Trim(pair[0], " \t\r\n")
		value := strings.Trim(strings.Trim(pair[1], " \t\r\n"), `"`)

		if _, keyExists := comp.Headers[key]; !keyExists {
			comp.Headers[key] = value
		}
	}

	// 5. Extract content-type if set on the second line
	for _, l := range lines[1:] {
		if strings.HasPrefix(l, "Content-Type: ") {
			comp.ContentType = strings.Trim(l[len("Content-Type: "):], " \t\r")
			break
		}

	}

	return nil
}

// GetHeader returns the header value associated with a key.
func (comp *Component) GetHeader(key string) string {
	value, ok := comp.Headers[key]

	if !ok {
		return ""
	}

	return value
}

// read all until the next boundary is found (and parse current MultipartData)
func (comp *Component) read(r *bufio.Reader, boundary string) error {

	headerRead := false
	rawHeader := make([]byte, 0)

	for { // Read until boundary or error

		line, err := r.ReadBytes('\n')

		// 1. Stop on error
		if err != nil {
			// remove last CR (newline)
			if strings.HasSuffix(string(comp.Data), "\n") {
				comp.Data = comp.Data[0 : len(comp.Data)-1]
			}
			if strings.HasSuffix(string(comp.Data), "\r") {
				comp.Data = comp.Data[0 : len(comp.Data)-1]
			}
			return err
		}

		// 2. Stop at boundary
		if strings.HasPrefix(string(line), boundary) {

			// remove last CR (newline)
			if strings.HasSuffix(string(comp.Data), "\n") {
				comp.Data = comp.Data[0 : len(comp.Data)-1]
			}
			if strings.HasSuffix(string(comp.Data), "\r") {
				comp.Data = comp.Data[0 : len(comp.Data)-1]
			}

			// io.EOF if last boundary
			if strings.Trim(string(line), " \t\r\n") == fmt.Sprintf("%s--", boundary) {
				return io.EOF
			}

			return nil
		}

		// 3. Ignore empty lines
		if len(strings.Trim(string(line), " \t\r\n")) > 0 {

			// add to header if not finished
			if !headerRead {
				rawHeader = append(rawHeader, line...)
				// else add to data (body)
			} else {
				comp.Data = append(comp.Data, line...)
			}

		} else if !headerRead { // if empty line, header has been read
			headerRead = true
			// rawHeader = append(rawHeader, line...)
			if err := comp.parseHeaders(rawHeader); err != nil {
				return err
			}
		}

	}

}
