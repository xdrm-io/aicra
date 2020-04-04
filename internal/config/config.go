package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"git.xdrm.io/go/aicra/datatype"
)

// Server definition
type Server struct {
	Types    []datatype.T
	Services []*Service
}

// Parse a reader into a server. Server.Types must be set beforehand to
// make datatypes available when checking and formatting the read configuration.
func (srv *Server) Parse(r io.Reader) error {
	if err := json.NewDecoder(r).Decode(&srv.Services); err != nil {
		return fmt.Errorf("%s: %w", ErrRead, err)
	}

	if err := srv.validate(); err != nil {
		return fmt.Errorf("%s: %w", ErrFormat, err)
	}

	return nil
}

// validate implements the validator interface
func (server Server) validate(datatypes ...datatype.T) error {
	for _, service := range server.Services {
		err := service.validate(server.Types...)
		if err != nil {
			return fmt.Errorf("%s '%s': %w", service.Method, service.Pattern, err)
		}
	}

	// check for collisions
	if err := server.collide(); err != nil {
		return fmt.Errorf("%s: %w", ErrFormat, err)
	}

	return nil
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

// SplitURL without empty sets
func SplitURL(url string) []string {
	trimmed := strings.Trim(url, " /\t\r\n")
	split := strings.Split(trimmed, "/")

	// remove empty set when empty url
	if len(split) == 1 && len(split[0]) == 0 {
		return []string{}
	}
	return split
}
