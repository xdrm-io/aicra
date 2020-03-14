package config

import (
	"fmt"
	"net/http"
	"strings"
)

// Match returns if this service would handle this HTTP request
func (svc *Service) Match(req *http.Request) bool {
	return false
}

func (svc *Service) checkMethod() error {
	for _, available := range availableHTTPMethods {
		if svc.Method == available {
			return nil
		}
	}
	return ErrUnknownMethod
}

func (svc *Service) checkPattern() error {
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

	// check capturing braces
	depth := 0
	for c, l := 1, length; c < l; c++ {
		char := svc.Pattern[c]

		if char == '{' {
			// opening brace when already opened
			if depth != 0 {
				return ErrInvalidPatternOpeningBrace
			}

			// not directly preceded by a slash
			if svc.Pattern[c-1] != '/' {
				return ErrInvalidPatternBracePosition
			}
			depth++
		}
		if char == '}' {
			// closing brace when already closed
			if depth != 1 {
				return ErrInvalidPatternClosingBrace
			}
			// not directly followed by a slash or end of pattern
			if c+1 < l && svc.Pattern[c+1] != '/' {
				return ErrInvalidPatternBracePosition
			}
			depth--
		}
	}

	return nil
}

func (svc *Service) checkAndFormatInput() error {

	// ignore no parameter
	if svc.Input == nil || len(svc.Input) < 1 {
		svc.Input = make(map[string]*Parameter, 0)
		return nil
	}

	// for each parameter
	for paramName, param := range svc.Input {

		// fail on invalid name
		if strings.Trim(paramName, "_") != paramName {
			return fmt.Errorf("%s: %w", paramName, ErrIllegalParamName)
		}

		// use param name if no rename
		if len(param.Rename) < 1 {
			param.Rename = paramName
		}

		err := param.checkAndFormat()
		if err != nil {
			return fmt.Errorf("%s: %w", paramName, err)
		}

		// check for name/rename conflict
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
