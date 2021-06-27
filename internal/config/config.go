package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/xdrm-io/aicra/validator"
)

// Server definition
type Server struct {
	Validators []validator.Type
	Services   []*Service
}

// Parse a configuration into a server. Server.Validators must be set beforehand
// to make datatypes available when checking and formatting the configuration.
func (s *Server) Parse(r io.Reader) error {
	err := json.NewDecoder(r).Decode(&s.Services)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrRead, err)
	}

	err = s.validate()
	if err != nil {
		return fmt.Errorf("%s: %w", ErrFormat, err)
	}
	return nil
}

// validate all services
func (s Server) validate(datatypes ...validator.Type) error {
	for _, service := range s.Services {
		err := service.validate(s.Validators...)
		if err != nil {
			return fmt.Errorf("%s '%s': %w", service.Method, service.Pattern, err)
		}
	}

	if err := s.collide(); err != nil {
		return fmt.Errorf("%s: %w", ErrFormat, err)
	}
	return nil
}

// Find a service matching an incoming HTTP request
func (s Server) Find(r *http.Request) *Service {
	for _, service := range s.Services {
		if matches := service.Match(r); matches {
			return service
		}
	}
	return nil
}

// collide returns if there is collision between any service for the same method
// and colliding paths. Note that service path collision detection relies on
// validators:
//  - example 1: `/user/{id}` and `/user/articles` will not collide as {id} is
//    an int and "articles" is not
//  - example 2: `/user/{name}` and `/user/articles` will collide as {name} is
//    a string so as "articles"
//  - example 3: `/user/{name}` and `/user/{id}` will collide as {name} and {id}
//    cannot be checked against their potential values
func (s *Server) collide() error {
	length := len(s.Services)

	// for each service combination
	for a := 0; a < length; a++ {
		for b := a + 1; b < length; b++ {
			aService := s.Services[a]
			bService := s.Services[b]

			if aService.Method != bService.Method {
				continue
			}

			aURIParts := SplitURL(aService.Pattern)
			bURIParts := SplitURL(bService.Pattern)
			if len(aURIParts) != len(bURIParts) {
				continue
			}

			err := checkURICollision(aURIParts, bURIParts, aService.Input, bService.Input)
			if err != nil {
				return fmt.Errorf("(%s '%s') vs (%s '%s'): %w", aService.Method, aService.Pattern, bService.Method, bService.Pattern, err)
			}
		}
	}

	return nil
}

// check if uri of services A and B collide
func checkURICollision(uriA, uriB []string, inputA, inputB map[string]*Parameter) error {
	var errors = []error{}

	// for each part
	for pi, aPart := range uriA {
		bPart := uriB[pi]

		// no need for further check as it has been done earlier in the validation process
		aIsCapture := len(aPart) > 1 && aPart[0] == '{'
		bIsCapture := len(bPart) > 1 && bPart[0] == '{'

		// both captures -> as we cannot check, consider a collision
		if aIsCapture && bIsCapture {
			errors = append(errors, fmt.Errorf("%w (path %s and %s)", ErrPatternCollision, aPart, bPart))
			continue
		}

		// no capture -> check strict equality
		if !aIsCapture && !bIsCapture {
			if aPart == bPart {
				errors = append(errors, fmt.Errorf("%w (same path '%s')", ErrPatternCollision, aPart))
				continue
			}
		}

		// A captures B -> check type (B is A ?)
		if aIsCapture {
			input, exists := inputA[aPart]

			// fail if no type or no validator
			if !exists || input.Validator == nil {
				errors = append(errors, fmt.Errorf("%w (invalid type for %s)", ErrPatternCollision, aPart))
				continue
			}

			// fail if not valid
			if _, valid := input.Validator(bPart); valid {
				errors = append(errors, fmt.Errorf("%w (%s captures '%s')", ErrPatternCollision, aPart, bPart))
				continue
			}

			// B captures A -> check type (A is B ?)
		} else if bIsCapture {
			input, exists := inputB[bPart]

			// fail if no type or no validator
			if !exists || input.Validator == nil {
				errors = append(errors, fmt.Errorf("%w (invalid type for %s)", ErrPatternCollision, bPart))
				continue
			}

			// fail if not valid
			if _, valid := input.Validator(aPart); valid {
				errors = append(errors, fmt.Errorf("%w (%s captures '%s')", ErrPatternCollision, bPart, aPart))
				continue
			}
		}

		errors = append(errors, nil)

	}

	// at least 1 URI part not matching -> no collision
	var firstError error
	for _, err := range errors {
		if err != nil && firstError == nil {
			firstError = err
		}

		if err == nil {
			return nil
		}
	}

	return firstError
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
