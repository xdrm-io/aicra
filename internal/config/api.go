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

// API definition
type API struct {
	Endpoints []*Endpoint `json:"endpoints"`
}

// Parse a configuration into a server. Server.Validators must be set beforehand
// to make datatypes available when checking and formatting the configuration.
func (s *API) Parse(r io.Reader) error {
	err := json.NewDecoder(r).Decode(&s.Endpoints)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrRead, err)
	}

	err = s.validate()
	if err != nil {
		return fmt.Errorf("%s: %w", ErrFormat, err)
	}
	return nil
}

// validate all endpoints
func (s API) validate() error {
	for _, endpoint := range s.Endpoints {
		err := endpoint.validate(s.Input, s.Output)
		if err != nil {
			return fmt.Errorf("%s %q: %w", endpoint.Method, endpoint.Pattern, err)
		}
	}

	if err := s.collide(); err != nil {
		return fmt.Errorf("%s: %w", ErrFormat, err)
	}
	return nil
}

// Find a endpoint matching an incoming HTTP request
func (s API) Find(r *http.Request) *Endpoint {
	for _, endpoint := range s.Endpoints {
		if matches := endpoint.Match(r); matches {
			return endpoint
		}
	}
	return nil
}

// collide returns if there is collision between any endpoint for the same method
// and colliding paths. Note that endpoint path collision detection relies on
// validators:
//   - example 1: `/user/{id}` and `/user/articles` will not collide as {id} is
//     an int and "articles" is not
//   - example 2: `/user/{name}` and `/user/articles` will collide as {name} is
//     a string so as "articles"
//   - example 3: `/user/{name}` and `/user/{id}` will collide as {name} and {id}
//     cannot be checked against their potential values
func (s *API) collide() error {
	length := len(s.Endpoints)

	// for each endpoint combination
	for a := 0; a < length; a++ {
		for b := a + 1; b < length; b++ {
			aendpoint := s.Endpoints[a]
			bendpoint := s.Endpoints[b]

			if aendpoint.Method != bendpoint.Method {
				continue
			}

			aURIParts := SplitURI(aendpoint.Pattern)
			bURIParts := SplitURI(bendpoint.Pattern)
			if len(aURIParts) != len(bURIParts) {
				continue
			}

			err := checkURICollision(aURIParts, bURIParts, aendpoint.Input, bendpoint.Input)
			if err != nil {
				return fmt.Errorf("(%s %q) vs (%s %q): %w", aendpoint.Method, aendpoint.Pattern, bendpoint.Method, bendpoint.Pattern, err)
			}
		}
	}

	return nil
}

// check if uri of endpoints A and B collide
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

// SplitURI without empty sets
func SplitURI(uri string) []string {
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
