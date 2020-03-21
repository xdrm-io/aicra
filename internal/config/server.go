package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"git.xdrm.io/go/aicra/datatype"
)

// Parse builds a server configuration from a json reader and checks for most format errors.
// you can provide additional DataTypes as variadic arguments
func Parse(r io.Reader, dtypes ...datatype.T) (*Server, error) {
	server := &Server{
		Types:    make([]datatype.T, 0),
		Services: make([]*Service, 0),
	}
	// add data types
	for _, dtype := range dtypes {
		server.Types = append(server.Types, dtype)
	}

	// parse JSON
	if err := json.NewDecoder(r).Decode(&server.Services); err != nil {
		return nil, fmt.Errorf("%s: %w", ErrRead, err)
	}

	// check services
	if err := server.checkAndFormat(); err != nil {
		return nil, fmt.Errorf("%s: %w", ErrFormat, err)
	}

	// check collisions
	if err := server.collide(); err != nil {
		return nil, fmt.Errorf("%s: %w", ErrFormat, err)
	}

	return server, nil
}

// Find a service matching an incoming HTTP request
func (server Server) Find(r *http.Request) *Service {
	for _, service := range server.Services {
		if matches := service.Match(r); matches {
			return service
		}
	}

	return nil
}

// collide returns if there is collision between services
func (server *Server) collide() error {
	length := len(server.Services)

	// for each service combination
	for a := 0; a < length; a++ {
		for b := a + 1; b < length; b++ {
			aService := server.Services[a]
			bService := server.Services[b]

			// ignore different method
			if aService.Method != bService.Method {
				continue
			}

			aParts := SplitURL(aService.Pattern)
			bParts := SplitURL(bService.Pattern)

			// not same size
			if len(aParts) != len(bParts) {
				continue
			}

			partErrors := make([]error, 0)

			// for each part
			for pi, aPart := range aParts {
				bPart := bParts[pi]

				aIsCapture := len(aPart) > 1 && aPart[0] == '{'
				bIsCapture := len(bPart) > 1 && bPart[0] == '{'

				// both captures -> as we cannot check, consider a collision
				if aIsCapture && bIsCapture {
					partErrors = append(partErrors, fmt.Errorf("(%s '%s') vs (%s '%s'): %w (path %s and %s)", aService.Method, aService.Pattern, bService.Method, bService.Pattern, ErrPatternCollision, aPart, bPart))
					continue
				}

				// no capture -> check equal
				if !aIsCapture && !bIsCapture {
					if aPart == bPart {
						partErrors = append(partErrors, fmt.Errorf("(%s '%s') vs (%s '%s'): %w (same path '%s')", aService.Method, aService.Pattern, bService.Method, bService.Pattern, ErrPatternCollision, aPart))
						continue
					}
				}

				// A captures B -> check type (B is A ?)
				if aIsCapture {
					input, exists := aService.Input[aPart]

					// fail if no type or no validator
					if !exists || input.Validator == nil {
						partErrors = append(partErrors, fmt.Errorf("(%s '%s') vs (%s '%s'): %w (invalid type for %s)", aService.Method, aService.Pattern, bService.Method, bService.Pattern, ErrPatternCollision, aPart))
						continue
					}

					// fail if not valid
					if _, valid := input.Validator(bPart); valid {
						partErrors = append(partErrors, fmt.Errorf("(%s '%s') vs (%s '%s'): %w (%s captures '%s')", aService.Method, aService.Pattern, bService.Method, bService.Pattern, ErrPatternCollision, aPart, bPart))
						continue
					}

					// B captures A -> check type (A is B ?)
				} else if bIsCapture {
					input, exists := bService.Input[bPart]

					// fail if no type or no validator
					if !exists || input.Validator == nil {
						partErrors = append(partErrors, fmt.Errorf("(%s '%s') vs (%s '%s'): %w (invalid type for %s)", aService.Method, aService.Pattern, bService.Method, bService.Pattern, ErrPatternCollision, bPart))
						continue
					}

					// fail if not valid
					if _, valid := input.Validator(aPart); valid {
						partErrors = append(partErrors, fmt.Errorf("(%s '%s') vs (%s '%s'): %w (%s captures '%s')", aService.Method, aService.Pattern, bService.Method, bService.Pattern, ErrPatternCollision, bPart, aPart))
						continue
					}
				}

				partErrors = append(partErrors, nil)

			}

			// if at least 1 url part does not match -> ok
			var firstError error
			oneMismatch := false
			for _, err := range partErrors {
				if err != nil && firstError == nil {
					firstError = err
				}

				if err == nil {
					oneMismatch = true
					continue
				}
			}

			if !oneMismatch {
				return firstError
			}

		}
	}

	return nil
}

// checkAndFormat checks for errors and missing fields and sets default values for optional fields.
func (server Server) checkAndFormat() error {
	for _, service := range server.Services {

		// check method
		err := service.checkMethod()
		if err != nil {
			return fmt.Errorf("%s '%s' [method]: %w", service.Method, service.Pattern, err)
		}

		// check pattern
		service.Pattern = strings.Trim(service.Pattern, " \t\r\n")
		err = service.checkPattern()
		if err != nil {
			return fmt.Errorf("%s '%s' [path]: %w", service.Method, service.Pattern, err)
		}

		// check description
		if len(strings.Trim(service.Description, " \t\r\n")) < 1 {
			return fmt.Errorf("%s '%s' [description]: %w", service.Method, service.Pattern, ErrMissingDescription)
		}

		// check input parameters
		err = service.checkAndFormatInput(server.Types)
		if err != nil {
			return fmt.Errorf("%s '%s' [in]: %w", service.Method, service.Pattern, err)
		}

		// fail if a brace capture remains undefined
		for _, capture := range service.Captures {
			if capture.Ref == nil {
				return fmt.Errorf("%s '%s' [in]: %s: %w", service.Method, service.Pattern, capture.Name, ErrUndefinedBraceCapture)
			}
		}

	}
	return nil
}
