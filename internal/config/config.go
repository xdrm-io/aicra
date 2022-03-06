package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/xdrm-io/aicra/validator"
)

// Server definition
type Server struct {
	// Input type validators available
	Input []validator.Type
	// Output types (no-op) validators available
	Output   []validator.Type
	Services []*Service
}

// AddInputValidator makes a new type available for services "in". It must be
// called before Parse() or will be ignored
func (s *Server) AddInputValidator(v validator.Type) {
	if s.Input == nil {
		s.Input = make([]validator.Type, 0)
	}
	s.Input = append(s.Input, v)
}

// AddOutputValidator adds an available output no-op validator for services "out"
// It only features a type name and a go type ; it must be called before Parse()
// or will be ignored
func (s *Server) AddOutputValidator(typename string, goType reflect.Type) {
	if s.Output == nil {
		s.Output = make([]validator.Type, 0)
	}
	s.Output = append(s.Output, noOp{name: typename, goType: goType})
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
func (s Server) validate() error {
	for _, service := range s.Services {
		err := service.validate(s.Input, s.Output)
		if err != nil {
			return fmt.Errorf("%s %q: %w", service.Method, service.Pattern, err)
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
				return fmt.Errorf("(%s %q) vs (%s %q): %w", aService.Method, aService.Pattern, bService.Method, bService.Pattern, err)
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
				errors = append(errors, fmt.Errorf("%w (same path %q)", ErrPatternCollision, aPart))
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
				errors = append(errors, fmt.Errorf("%w (%s captures %q)", ErrPatternCollision, aPart, bPart))
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
				errors = append(errors, fmt.Errorf("%w (%s captures %q)", ErrPatternCollision, bPart, aPart))
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

// noOp defines a no-op validator used for output parameters
type noOp struct {
	name   string
	goType reflect.Type
}

func (v noOp) GoType() reflect.Type {
	return v.goType
}
func (v noOp) Validator(typename string, avail ...validator.Type) validator.ValidateFunc {
	if typename != v.name {
		return nil
	}
	return func(value interface{}) (interface{}, bool) {
		return reflect.Zero(v.goType), false
	}
}
