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

var importNameRe = regexp.MustCompile(`^[a-z0-9_]+$`)

// validate the configuration
func (s API) validate() error {
	if len(s.Package) == 0 {
		return ErrPackageMissing
	}

	var builtin = []string{"fmt", "context", "http", "aicra", "builtin", "runtime"}
	var uniqPath = map[string]struct{}{}
	for alias, path := range s.Imports {
		if !importNameRe.MatchString(alias) {
			return fmt.Errorf("import '%s': %w", alias, ErrImportCharset)
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
func (s API) Find(r *http.Request) *Endpoint {
	for _, endpoint := range s.Endpoints {
		if matches := endpoint.Match(r); matches {
			return endpoint
		}
	}
	return nil
}

// Validate the config with code-generated validators
func (s API) Validate(avail Validators) error {
	for _, endpoint := range s.Endpoints {
		err := endpoint.Validate(avail)
		if err != nil {
			return fmt.Errorf("%s %q: %w", endpoint.Method, endpoint.Pattern, err)
		}
	}
	return nil
}

// URIFragments splits an uri into fragments with removing empty sets
func URIFragments(uri string) []string {
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
