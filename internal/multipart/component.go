package multipart

import (
	"bufio"
	"strings"
)

func (comp *Component) parseHeaders(_raw []byte) error {

	// 1. Extract lines
	_lines := strings.Split(string(_raw), "\n")
	if len(_lines) < 2 {
		return ErrNoHeader
	}

	// 2. trim each line + remove 'Content-Disposition' prefix
	trimmed := strings.Trim(_lines[0], " \t")
	header := trimmed

	if !strings.HasPrefix(trimmed, "Content-Disposition: form-data;") {
		return ErrNoHeader
	}
	header = strings.Trim(trimmed[len("Content-Disposition: form-data;"):], " \t")

	if len(header) < 1 {
		return ErrNoHeader
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
	for _, l := range _lines[1:] {

		if strings.HasPrefix(l, "Content-Type: ") {
			comp.ContentType = strings.Trim(l[len("Content-Type: "):], " \t")
			break
		}

	}

	return nil

}

// GetHeader returns the header value associated with a key, empty string if not found
func (comp *Component) GetHeader(_key string) string {
	value, ok := comp.Headers[_key]

	if !ok {
		return ""
	}

	return value
}

// read all until the next boundary is found (and parse current MultipartData)
func (comp *Component) read(_reader *bufio.Reader, _boundary string) error {

	headerRead := false
	rawHeader := make([]byte, 0)

	for { // Read until boundary or error

		line, err := _reader.ReadBytes('\n')

		// 1. Stop on error
		if err != nil {
			// remove last CR (newline)
			if string(comp.Data[len(comp.Data)-1]) == "\n" {
				comp.Data = comp.Data[0 : len(comp.Data)-1]
			}
			return err
		}

		// 2. Stop at boundary
		if strings.HasPrefix(string(line), _boundary) {

			// remove last CR (newline)
			if string(comp.Data[len(comp.Data)-1]) == "\n" {
				comp.Data = comp.Data[0 : len(comp.Data)-1]
			}
			return nil
		}

		// 3. Ignore empty lines
		if string(line) != "\n" && len(line) > 0 {

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