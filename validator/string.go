package validator

import (
	"regexp"
	"strconv"
)

var (
	fixedLengthRegex    = regexp.MustCompile(`^string\((\d+)\)$`)
	variableLengthRegex = regexp.MustCompile(`^string\((\d+), ?(\d+)\)$`)
)

// String makes the types beloz available in the aicra configuration:
// - "string" considers any string valid
// - "string(n)" considers any string with an exact size of `n` valid
// - "string(a,b)" considers any string with a size between `a` and `b` valid
// > for the last one, `a` and `b` are included in the valid sizes
type String struct{}

// Validate implements Validator for strings with any/fixed/bound sizes
func (s String) Validate(typename string) ExtractFunc[string] {
	var (
		simple      = (typename == "string")
		fixedLength = fixedLengthRegex.FindStringSubmatch(typename)
		varLength   = variableLengthRegex.FindStringSubmatch(typename)
	)

	// ignore unknown typename
	if !simple && fixedLength == nil && varLength == nil {
		return nil
	}

	var min, max int
	if fixedLength != nil {
		exLen, ok := s.getFixedLength(fixedLength)
		if !ok {
			return nil
		}
		min = exLen
		max = exLen
	}
	if varLength != nil {
		exMin, exMax, ok := s.getVariableLength(varLength)
		if !ok {
			return nil
		}
		min = exMin
		max = exMax
	}

	return func(value interface{}) (string, bool) {
		str, isStr := value.(string)
		bytes, isBytes := value.([]byte)
		if isBytes {
			str = string(bytes)
			isStr = true
		}

		if !isStr {
			return "", false
		}

		if simple {
			return str, true
		}

		// check length against previously extracted length
		l := len(str)
		return str, l >= min && l <= max
	}
}

// getFixedLength returns the fixed length from regex matches and a success state.
func (String) getFixedLength(regexMatches []string) (int, bool) {
	if len(regexMatches) < 2 {
		return 0, false
	}
	fixedLength, err := strconv.ParseInt(regexMatches[1], 10, 64)
	if err != nil || fixedLength < 0 {
		return 0, false
	}

	return int(fixedLength), true
}

// getVariableLength returns the length min and max from regex matches and a success state.
func (String) getVariableLength(regexMatches []string) (int, int, bool) {
	if len(regexMatches) < 3 {
		return 0, 0, false
	}
	minLen, err := strconv.ParseInt(regexMatches[1], 10, 64)
	if err != nil || minLen < 0 {
		return 0, 0, false
	}
	maxLen, err := strconv.ParseInt(regexMatches[2], 10, 64)
	if err != nil || maxLen < 0 {
		return 0, 0, false
	}
	return int(minLen), int(maxLen), true
}
