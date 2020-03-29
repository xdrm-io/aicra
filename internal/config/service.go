package config

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"git.xdrm.io/go/aicra/datatype"
)

var braceRegex = regexp.MustCompile(`^{([a-z_-]+)}$`)
var queryRegex = regexp.MustCompile(`^GET@([a-z_-]+)$`)

// Match returns if this service would handle this HTTP request
func (svc *Service) Match(req *http.Request) bool {
	// method
	if req.Method != svc.Method {
		return false
	}

	// check path
	if !svc.matchPattern(req.RequestURI) {
		return false
	}

	// check and extract input
	// todo: check if input match and extract models

	return true
}

// checks if an uri matches the service's pattern
func (svc *Service) matchPattern(uri string) bool {
	uriparts := SplitURL(uri)
	parts := SplitURL(svc.Pattern)

	// fail if size differ
	if len(uriparts) != len(parts) {
		return false
	}

	// root url '/'
	if len(parts) == 0 {
		return true
	}

	// check part by part
	for i, part := range parts {
		uripart := uriparts[i]

		isCapture := len(part) > 0 && part[0] == '{'

		// if no capture -> check equality
		if !isCapture {
			if part != uripart {
				return false
			}
			continue
		}

		param, exists := svc.Input[part]

		// fail if no validator
		if !exists || param.Validator == nil {
			return false
		}

		// fail if not type-valid
		if _, valid := param.Validator(uripart); !valid {
			return false
		}
	}

	return true
}

// Validate implements the validator interface
func (svc *Service) Validate(datatypes ...datatype.T) error {
	// check method
	err := svc.isMethodAvailable()
	if err != nil {
		return fmt.Errorf("field 'method': %w", err)
	}

	// check pattern
	svc.Pattern = strings.Trim(svc.Pattern, " \t\r\n")
	err = svc.isPatternValid()
	if err != nil {
		return fmt.Errorf("field 'path': %w", err)
	}

	// check description
	if len(strings.Trim(svc.Description, " \t\r\n")) < 1 {
		return fmt.Errorf("field 'description': %w", ErrMissingDescription)
	}

	// check input parameters
	err = svc.validateInput(datatypes)
	if err != nil {
		return fmt.Errorf("field 'in': %w", err)
	}

	// fail if a brace capture remains undefined
	for _, capture := range svc.Captures {
		if capture.Ref == nil {
			return fmt.Errorf("field 'in': %s: %w", capture.Name, ErrUndefinedBraceCapture)
		}
	}

	return nil
}

func (svc *Service) isMethodAvailable() error {
	for _, available := range availableHTTPMethods {
		if svc.Method == available {
			return nil
		}
	}
	return ErrUnknownMethod
}

func (svc *Service) isPatternValid() error {
	length := len(svc.Pattern)

	// empty pattern
	if length < 1 {
		return ErrInvalidPattern
	}

	if length > 1 {
		// pattern not starting with '/' or ending with '/'
		if svc.Pattern[0] != '/' || svc.Pattern[length-1] == '/' {
			return ErrInvalidPattern
		}
	}

	// for each slash-separated chunk
	parts := SplitURL(svc.Pattern)
	for i, part := range parts {
		if len(part) < 1 {
			return ErrInvalidPattern
		}

		// if brace capture
		if matches := braceRegex.FindAllStringSubmatch(part, -1); len(matches) > 0 && len(matches[0]) > 1 {
			braceName := matches[0][1]

			// append
			if svc.Captures == nil {
				svc.Captures = make([]*BraceCapture, 0)
			}
			svc.Captures = append(svc.Captures, &BraceCapture{
				Index: i,
				Name:  braceName,
				Ref:   nil,
			})
			continue
		}

		// fail on invalid format
		if strings.ContainsAny(part, "{}") {
			return ErrInvalidPatternBraceCapture
		}

	}

	return nil
}

func (svc *Service) validateInput(types []datatype.T) error {

	// ignore no parameter
	if svc.Input == nil || len(svc.Input) < 1 {
		svc.Input = make(map[string]*Parameter, 0)
		return nil
	}

	// for each parameter
	for paramName, param := range svc.Input {
		if len(paramName) < 1 {
			return fmt.Errorf("%s: %w", paramName, ErrIllegalParamName)
		}

		// fail if brace capture does not exists in pattern
		var iscapture, isquery bool
		if matches := braceRegex.FindAllStringSubmatch(paramName, -1); len(matches) > 0 && len(matches[0]) > 1 {
			braceName := matches[0][1]

			found := false
			for _, capture := range svc.Captures {
				if capture.Name == braceName {
					capture.Ref = param
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("%s: %w", paramName, ErrUnspecifiedBraceCapture)
			}
			iscapture = true

		} else if matches := queryRegex.FindAllStringSubmatch(paramName, -1); len(matches) > 0 && len(matches[0]) > 1 {

			queryName := matches[0][1]

			// init map
			if svc.Query == nil {
				svc.Query = make(map[string]*Parameter)
			}
			svc.Query[queryName] = param
			isquery = true
		} else {
			if svc.Form == nil {
				svc.Form = make(map[string]*Parameter)
			}
			svc.Form[paramName] = param
		}

		// fail if capture or query without rename
		if len(param.Rename) < 1 && (iscapture || isquery) {
			return fmt.Errorf("%s: %w", paramName, ErrMandatoryRename)
		}

		// use param name if no rename
		if len(param.Rename) < 1 {
			param.Rename = paramName
		}

		err := param.Validate(types...)
		if err != nil {
			return fmt.Errorf("%s: %w", paramName, err)
		}

		// capture parameter cannot be optional
		if iscapture && param.Optional {
			return fmt.Errorf("%s: %w", paramName, ErrIllegalOptionalURIParam)
		}

		// fail on name/rename conflict
		for paramName2, param2 := range svc.Input {
			// ignore self
			if paramName == paramName2 {
				continue
			}

			// 3.2.1. Same rename field
			// 3.2.2. Not-renamed field matches a renamed field
			// 3.2.3. Renamed field matches name
			if param.Rename == param2.Rename || paramName == param2.Rename || paramName2 == param.Rename {
				return fmt.Errorf("%s: %w", paramName, ErrParamNameConflict)
			}

		}

	}

	return nil
}
