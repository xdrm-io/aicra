package multipart

import (
	"fmt"
	"strings"
)

// Read all until the next boundary is found
func (i *Reader) readComponent() ([]string, error) {

	component := make([]string, 0)

	for { // Read until boundary or error

		line, _, err := i.reader.ReadLine()

		/* (1) Stop on error */
		if err != nil {
			return component, err
		}

		/* (2) Stop at boundary */
		if strings.HasPrefix(string(line), i.boundary) {
			return component, err
		}

		/* (3) Ignore empty lines */
		if len(line) > 0 {
			component = append(component, string(line))
		}

	}

}

// Parses a single component from its raw lines
func (i *Reader) parseComponent(line []string) error {

	// next line index to use
	cursor := 1

	/* (1) Fail if invalid line count */
	if len(line) < 2 {
		return fmt.Errorf("Missing data to parse component")
	}

	/* (2) Split meta data */
	meta := strings.Split(line[0], "; ")

	if len(meta) < 2 {
		return fmt.Errorf("Missing component meta data")
	}

	/* (3) Extract name */
	if !strings.HasPrefix(meta[1], `name="`) {
		return fmt.Errorf("Cannot extract component name")
	}
	name := meta[1][len(`name="`) : len(meta[1])-1]

	/* (4) Check if it is a file */
	isFile := len(meta) > 2 && strings.HasPrefix(meta[2], `filename="`)

	// skip next line (Content-Type) if file
	if isFile {
		cursor++
	}

	/* (5) Create index if name not already used */
	already, isset := i.Components[name]
	if !isset {

		i.Components[name] = &Component{
			File: isFile,
			Data: make([]string, 0),
		}
		already = i.Components[name]

	}

	/* (6) Store new value */
	already.Data = append(already.Data, strings.Join(line[cursor:], "\n"))

	return nil
}
