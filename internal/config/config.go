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
func checkURICollision(aURI, bURI []string, aInput, bInput map[string]*Parameter) error {
	var err error

	// for each part
	for i, aSeg := range aURI {
		var (
			bSeg = bURI[i]

			// no need for check deeper as it has been done earlier in the
			// validation process
			aIsCapture = len(aSeg) > 1 && aSeg[0] == '{'
			bIsCapture = len(bSeg) > 1 && bSeg[0] == '{'
		)

		// both captures -> as we cannot check, consider a collision
		if aIsCapture && bIsCapture {
			err = fmt.Errorf("%w (path %s and %s)", ErrPatternCollision, aSeg, bSeg)
			continue
		}

		// no capture -> check strict equality
		if !aIsCapture && !bIsCapture && aSeg == bSeg {
			err = fmt.Errorf("%w (same path %q)", ErrPatternCollision, aSeg)
			continue
		}

		// A captures B -> fail if B is a valid A value
		if aIsCapture && validates(aInput, aSeg, bSeg) {
			err = fmt.Errorf("%w (%s captures %q)", ErrPatternCollision, aSeg, bSeg)
			continue
		}
		// B captures A -> fail is A is a valid B value
		if bIsCapture && validates(bInput, bSeg, aSeg) {
			err = fmt.Errorf("%w (%s captures %q)", ErrPatternCollision, bSeg, aSeg)
			continue
		}
		// no match for at least one segment -> no collision
		return nil
	}
	return err
}

// validates returns whether a parameter validates a given value
func validates(params map[string]*Parameter, checkerName, value string) bool {
	checker, exists := params[checkerName]
	if !exists || checker.Validator == nil {
		panic(fmt.Errorf("invalid validator %q", checkerName))
	}
	_, valid := checker.Validator(value)
	return valid
}

// SplitURL without empty sets
func SplitURL(uri string) []string {
	if len(uri) == 0 || uri == "/" {
		return []string{}
	}
	if len(uri) > 0 && uri[0] == '/' {
		uri = uri[1:]
	}
	if len(uri) > 0 && uri[len(uri)-1] == '/' {
		uri = uri[:len(uri)-1]
	}
	for len(uri) > 0 && uri[len(uri)-1] == '/' {
		uri = uri[:len(uri)-1]
	}
	return strings.Split(uri, "/")
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
		return reflect.Zero(v.goType), true
	}
}
