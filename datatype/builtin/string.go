package builtin

import (
	"reflect"
	"regexp"
	"strconv"

	"github.com/xdrm-io/aicra/datatype"
)

var fixedLengthRegex = regexp.MustCompile(`^string\((\d+)\)$`)
var variableLengthRegex = regexp.MustCompile(`^string\((\d+), ?(\d+)\)$`)

// StringDataType is what its name tells
type StringDataType struct{}

// Type returns the type of data
func (StringDataType) Type() reflect.Type {
	return reflect.TypeOf(string(""))
}

// Build returns the validator.
// availables type names are : `string`, `string(length)` and `string(minLength, maxLength)`.
func (s StringDataType) Build(typeName string, registry ...datatype.T) datatype.Validator {
	simple := typeName == "string"
	fixedLengthMatches := fixedLengthRegex.FindStringSubmatch(typeName)
	variableLengthMatches := variableLengthRegex.FindStringSubmatch(typeName)

	// nothing if type not handled
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
func (StringDataType) getFixedLength(regexMatches []string) (int, bool) {
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
func (StringDataType) getVariableLength(regexMatches []string) (int, int, bool) {
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
