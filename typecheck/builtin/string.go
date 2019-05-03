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
func (s String) Checker(typeName string) typecheck.CheckerFunc {
	isSimpleString := typeName == "string"
	fixedLengthMatches := fixedLengthRegex.FindStringSubmatch(typeName)
	variableLengthMatches := variableLengthRegex.FindStringSubmatch(typeName)

	// nothing if type not handled
	if !isSimpleString && fixedLengthMatches == nil && variableLengthMatches == nil {
		return nil
	}

	var (
		mustFail bool
		min, max int
	)

	// extract fixed length
	if fixedLengthMatches != nil {
		exLen, ok := s.getFixedLength(fixedLengthMatches)
		if !ok {
			mustFail = true
		}
		min = exLen
		max = exLen

		// extract variable length
	} else if variableLengthMatches != nil {
		exMin, exMax, ok := s.getVariableLength(variableLengthMatches)
		if !ok {
			mustFail = true
		}
		min = exMin
		max = exMax
	}

	return func(value interface{}) bool {
		// preprocessing error
		if mustFail {
			return false
		}

		// check type
		strValue, isString := value.(string)
		if !isString {
			return false
		}

		if isSimpleString {
			return true
		}

		// check length against previously extracted length
		l := len(strValue)
		return l >= min && l <= max
	}
}

// getFixedLength returns the fixed length from regex matches and a success state.
func (String) getFixedLength(regexMatches []string) (int, bool) {
	// incoherence error
	if regexMatches == nil || len(regexMatches) < 2 {
		return 0, false
	}

	// extract length
	fixedLength, err := strconv.ParseInt(regexMatches[1], 10, 64)
	if err != nil || fixedLength < 0 {
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
	minLen, err := strconv.ParseInt(regexMatches[1], 10, 64)
	if err != nil || minLen < 0 {
		return 0, 0, false
	}
	// extract maximum length
	maxLen, err := strconv.ParseInt(regexMatches[2], 10, 64)
	if err != nil || maxLen < 0 {
		return 0, 0, false
	}

	return int(minLen), int(maxLen), true
}
