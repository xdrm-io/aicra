package validator

import (
	"reflect"
	"regexp"
	"strconv"
)

var (
	fixedLengthRegex    = regexp.MustCompile(`^string\((\d+)\)$`)
	variableLengthRegex = regexp.MustCompile(`^string\((\d+), ?(\d+)\)$`)
)

// StringType makes the types beloz available in the aicra configuration:
// - "string" considers any string valid
// - "string(n)" considers any string with an exact size of `n` valid
// - "string(a,b)" considers any string with a size between `a` and `b` valid
// > for the last one, `a` and `b` are included in the valid sizes
type StringType struct{}

// GoType returns the `string` type
func (StringType) GoType() reflect.Type {
	return reflect.TypeOf(string(""))
}

// Validator for strings with any/fixed/bound sizes
func (s StringType) Validator(typename string, avail ...Type) ValidateFunc {
	var (
		simple                = (typename == "string")
		fixedLengthMatches    = fixedLengthRegex.FindStringSubmatch(typename)
		variableLengthMatches = variableLengthRegex.FindStringSubmatch(typename)
	)

	// ignore unknown typename
	if !simple && fixedLengthMatches == nil && variableLengthMatches == nil {
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
			return nil
		}
		min = exLen
		max = exLen

		// extract variable length
	} else if variableLengthMatches != nil {
		exMin, exMax, ok := s.getVariableLength(variableLengthMatches)
		if !ok {
			return nil
		}
		min = exMin
		max = exMax
	}

	return func(value interface{}) (interface{}, bool) {
		// preprocessing error
		if mustFail {
			return "", false
		}

		// check type
		strValue, isString := value.(string)
		byteSliceValue, isByteSlice := value.([]byte)
		if !isString && isByteSlice {
			strValue = string(byteSliceValue)
			isString = true
		}

		if !isString {
			return "", false
		}

		if simple {
			return strValue, true
		}

		// check length against previously extracted length
		l := len(strValue)
		return strValue, l >= min && l <= max
	}
}

// getFixedLength returns the fixed length from regex matches and a success state.
func (StringType) getFixedLength(regexMatches []string) (int, bool) {
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
func (StringType) getVariableLength(regexMatches []string) (int, int, bool) {
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
