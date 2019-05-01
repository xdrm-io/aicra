package builtin

import (
	"regexp"
	"strconv"

	"git.xdrm.io/go/aicra/typecheck"
)

var fixedLengthRegex = regexp.MustCompile(`^string\((\d+)\)$`)
var variableLengthRegex = regexp.MustCompile(`^string\((\d+), ?(\d+)\)$`)

// String checks if a value is a string
type String struct{}

// NewString returns a bare string type checker
func NewString() *String {
	return &String{}
}

// Checker returns the checker function. Availables type names are : `string`, `string(length)` and `string(minLength, maxLength)`.
func (s String) Checker(typeName string) typecheck.Checker {
	isSimpleString := typeName == "string"
	fixedLengthMatches := fixedLengthRegex.FindStringSubmatch(typeName)
	variableLengthMatches := variableLengthRegex.FindStringSubmatch(typeName)

	// nothing if type not handled
	if !isSimpleString && fixedLengthMatches == nil && variableLengthMatches == nil {
		return nil
	}

	return func(value interface{}) bool {
		// check type
		strValue, isString := value.(string)
		if !isString {
			return false
		}

		// check fixed length
		if fixedLengthMatches != nil {
			// incoherence fail
			if len(fixedLengthMatches) < 2 {
				return false
			}

			// extract length
			fixedLen, err := strconv.ParseUint(fixedLengthMatches[1], 10, 64)
			if err != nil {
				return false
			}

			// check against value
			return len(strValue) == int(fixedLen)
		}

		// check variable length
		if variableLengthMatches != nil {

			minLen, maxLen, ok := s.getVariableLength(variableLengthMatches)
			if !ok {
				return false
			}

			// check against value
			return len(strValue) >= minLen && len(strValue) <= maxLen
		}

		// if should NEVER be here ; so fail
		return false
	}
}

// getFixedLength returns the fixed length from regex matches and a success state.
func (String) getFixedLength(regexMatches []string) (int, bool) {
	// incoherence error
	if regexMatches == nil || len(regexMatches) < 2 {
		return 0, false
	}

	// extract length
	fixedLength, err := strconv.ParseUint(regexMatches[1], 10, 64)
	if err != nil {
		return 0, false
	}

	return int(fixedLength), true
}

// getVariableLength returns the length min and max from regex matches and a success state.
func (String) getVariableLength(regexMatches []string) (int, int, bool) {
	// incoherence error
	if regexMatches == nil || len(regexMatches) < 3 {
		return 0, 0, false
	}

	// extract minimum length
	minLen, err := strconv.ParseUint(regexMatches[1], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	// extract maximum length
	maxLen, err := strconv.ParseUint(regexMatches[2], 10, 64)
	if err != nil {
		return 0, 0, false
	}

	return int(minLen), int(maxLen), true
}
