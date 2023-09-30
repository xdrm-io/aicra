package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
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
	for _, endpoint := range s.Endpoints {
		if matches := endpoint.Match(r, validators); matches {
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
			return fmt.Errorf("%s %q: %w", endpoint.Method, endpoint.Pattern, err)
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
		captures[endpoint.Method+endpoint.Pattern] = captureSpec(endpoint)
	}

	length := len(s.Endpoints)

	// for each service combination
	for a := 0; a < length; a++ {
		aEndpoint := s.Endpoints[a]
		aCaptures := captures[aEndpoint.Method+aEndpoint.Pattern]

		for b := a + 1; b < length; b++ {
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
				return fmt.Errorf("(%s %q) vs (%s %q): %w", aEndpoint.Method, aEndpoint.Pattern, bEndpoint.Method, bEndpoint.Pattern, err)
			}
		}
	}

	return nil
}
