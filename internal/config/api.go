package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/xdrm-io/aicra/validator"
)

// API definition
type API struct {
	Package    string               `json:"package"`
	Imports    map[string]string    `json:"imports"`
	Validators map[string]Validator `json:"validators"`
	Endpoints  []*Endpoint          `json:"endpoints"`
}

// UnmarshalJSON with custom validation
func (s *API) UnmarshalJSON(b []byte) error {
	type receiver API
	var r receiver
	if err := json.Unmarshal(b, &r); err != nil {
		return err
	}

	s.Package = r.Package
	s.Imports = r.Imports
	s.Validators = r.Validators
	s.Endpoints = r.Endpoints
	return s.validate()
}

var importNameRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
var importPathRe = regexp.MustCompile(`^[a-zA-Z0-9_/\.-]+$`)

// validate the configuration
func (s API) validate() error {
	if len(s.Package) == 0 {
		return ErrPackageMissing
	}
	if len(s.Validators) == 0 {
		return ErrValidatorsMissing
	}
	if s.Endpoints == nil {
		return ErrEndpointsMissing
	}

	var builtin = []string{"fmt", "context", "http", "aicra", "builtin", "runtime"}
	var uniqPath = map[string]struct{}{}
	for alias, path := range s.Imports {
		if !importNameRe.MatchString(alias) {
			return fmt.Errorf("import '%s': %w", alias, ErrImportAliasCharset)
		}
		if !importPathRe.MatchString(path) {
			return fmt.Errorf("import '%s': %w", alias, ErrImportPathCharset)
		}
		for _, forbidden := range builtin {
			if alias == forbidden {
				return fmt.Errorf("import '%s': %w (%v)", alias, ErrImportReserved, builtin)
			}
		}
		if _, already := uniqPath[path]; already {
			return fmt.Errorf("import '%s': %w", alias, ErrImportTwice)
		}
		uniqPath[path] = struct{}{}
	}
	return nil
}

// Find a endpoint matching an incoming HTTP request
func (s API) Find(r *http.Request, validators Validators) *Endpoint {
	if r == nil {
		return nil
	}

	method, fragments := r.Method, URIFragments(r.URL.Path)
	for _, endpoint := range s.Endpoints {
		if matches := endpoint.Match(method, fragments, validators); matches {
			return endpoint
		}
	}
	return nil
}

// RuntimeCheck fails when the config is invalid with the code-generated
// validators
func (s API) RuntimeCheck(avail Validators) error {
	for _, endpoint := range s.Endpoints {
		err := endpoint.RuntimeCheck(avail)
		if err != nil {
			return fmt.Errorf("'%s %s': %w", endpoint.Method, endpoint.Pattern, err)
		}
	}
	return s.collide(avail)
}

// URIFragments splits an uri into fragments with removing empty sets
func URIFragments(uri string) []string {
	if len(uri) == 0 || uri == "/" {
		return []string{}
	}
	if len(uri) > 0 && uri[0] == '/' {
		uri = uri[1:]
	}
	for len(uri) > 0 && uri[len(uri)-1] == '/' {
		uri = uri[:len(uri)-1]
	}
	return strings.Split(uri, "/")
}

// collide returns if there is collision between any service for the same method
// and colliding paths. Note that service path collision detection relies on
// validators:
//   - example 1: `/user/{id}` and `/user/articles` will not collide as {id} is
//     an int and "articles" is not
//   - example 2: `/user/{name}` and `/user/articles` will collide as {name} is
//     a string so as "articles"
//   - example 3: `/user/{name}` and `/user/{id}` will collide as {name} and {id}
//     cannot be checked against their potential values
func (s API) collide(avail Validators) error {
	// process captures' validation specs for every endpoint
	captures := make(map[string]map[int]captureValidation, len(s.Endpoints))
	for _, endpoint := range s.Endpoints {
		cap, err := captureSpec(endpoint)
		if err != nil {
			return fmt.Errorf("'%s %s': %w", endpoint.Method, endpoint.Pattern, err)
		}
		captures[endpoint.Method+endpoint.Pattern] = cap
	}

	l := len(s.Endpoints)

	// for each service combination
	for a := 0; a < l; a++ {
		aEndpoint := s.Endpoints[a]
		aCaptures := captures[aEndpoint.Method+aEndpoint.Pattern]

		for b := a + 1; b < l; b++ {
			bEndpoint := s.Endpoints[b]
			bCaptures := captures[bEndpoint.Method+bEndpoint.Pattern]

			if aEndpoint.Method != bEndpoint.Method {
				continue
			}

			aFragments := URIFragments(aEndpoint.Pattern)
			bFragments := URIFragments(bEndpoint.Pattern)
			if len(aFragments) != len(bFragments) {
				continue
			}

			err := checkURICollision(aFragments, bFragments, aCaptures, bCaptures, avail)
			if err != nil {
				return fmt.Errorf("'%s %s' vs '%s %s': %w", aEndpoint.Method, aEndpoint.Pattern, bEndpoint.Method, bEndpoint.Pattern, err)
			}
		}
	}

	return nil
}

type captureValidation struct {
	ValidatorName   string
	ValidatorParams []string
}

// captures returns the captures' validators for an endpoint indexed by their
// index in the URI
func captureSpec(endpoint *Endpoint) (map[int]captureValidation, error) {
	captures := make(map[int]captureValidation, len(endpoint.Captures))
	for _, capture := range endpoint.Captures {
		p, ok := endpoint.Input[`{`+capture.Name+`}`]
		if !ok {
			return nil, fmt.Errorf("%w: %d %q", ErrBraceCaptureUndefined, capture.Index, capture.Name)
		}
		captures[capture.Index] = captureValidation{
			ValidatorName:   p.ValidatorName,
			ValidatorParams: p.ValidatorParams,
		}
	}
	return captures, nil
}

// check if uri of services A and B collide
func checkURICollision(aFragments, bFragments []string, aCaptures, bCaptures map[int]captureValidation, avail Validators) error {
	var err error

	// for each part
	for i, aSeg := range aFragments {
		var (
			bSeg = bFragments[i]

			aCapture, aIsCapture = aCaptures[i]
			bCapture, bIsCapture = bCaptures[i]
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
		if aIsCapture {
			extractFn := avail[aCapture.ValidatorName].Validate(aCapture.ValidatorParams)
			e := validates(extractFn, bSeg)
			if e != nil {
				err = fmt.Errorf("%w: %s captures %q", e, aSeg, bSeg)
				continue
			}
		}
		// B captures A -> fail is A is a valid B value
		if bIsCapture {
			extractFn := avail[bCapture.ValidatorName].Validate(bCapture.ValidatorParams)
			e := validates(extractFn, aSeg)
			if e != nil {
				err = fmt.Errorf("%w: %s captures %q", e, bSeg, aSeg)
				continue
			}
		}
		// no match for at least one segment -> no collision
		return nil
	}
	return err
}

// validates returns nil when the value is invalid for the validator
func validates(extractFn validator.ExtractFunc[any], value string) error {
	_, valid := extractFn(value)
	if valid {
		return ErrPatternCollision
	}
	return nil
}
