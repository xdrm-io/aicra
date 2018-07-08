package multipart

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Creates a new multipart reader from an http.Request
func CreateReader(req *http.Request) *MultipartReader {

	/* (1) extract boundary */
	boundary := req.Header.Get("Content-Type")[len("multipart/form-data; boundary="):]
	boundary = fmt.Sprintf("--%s", boundary)

	/* (2) init reader */
	i := &MultipartReader{
		reader:     bufio.NewReader(req.Body),
		boundary:   boundary,
		Components: make(map[string]*MultipartComponent),
	}

	/* (3) Place reader cursor after first boundary */
	var (
		err  error
		line []byte
	)

	for err == nil && string(line) != boundary {
		line, _, err = i.reader.ReadLine()
	}

	return i

}

// Parses the multipart components from the request
func (i *MultipartReader) Parse() error {

	/* (1) For each component (until boundary) */
	for {

		// 1. Read component
		component, err := i.readComponent()

		// 2. Stop at EOF
		if err == io.EOF {
			return nil
		}

		// 3. Dispatch error
		if err != nil {
			return err
		}

		// 4. parse component
		err = i.parseComponent(component)

		if err != nil {
			log.Printf("%s\n", err)
		}

	}

}
