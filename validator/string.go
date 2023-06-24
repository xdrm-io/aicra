package validator

import (
	"regexp"
	"strconv"
)

var (
	fixedLengthRegex    = regexp.MustCompile(`^string\((\d+)\)$`)
	variableLengthRegex = regexp.MustCompile(`^string\((\d+), ?(\d+)\)$`)
)

// String valides strings and can accept 0, 1 or 2 parameters. Only accepts
// uint32 as parameters
// * 0 param -> string of any size
// * 1 param -> string of exactly the specified size
// * 2 params -> string with a size between 1st and 2nd parameter
// (included)
//
// Example parameters :
// * {}  : string of any size
// * {1} : string of size 1
// * {1,3} : string of size between 1 and 3 (1, 2 or 3)
type String struct{}

// Validate implements Validator for strings with any/fixed/bound sizes
func (s String) Validate(params []string) ExtractFunc[string] {
	if len(params) > 2 {
		return nil
	}
	freeSize := len(params) == 0
	var min, max int

	if len(params) == 1 { // fixed size
		size, err := strconv.ParseUint(params[0], 10, 32)
		if err != nil {
			return nil
		}
		min, max = int(size), int(size)
	}

	if len(params) == 2 { // variable size
		umin, err := strconv.ParseUint(params[0], 10, 32)
		if err != nil {
			return nil
		}
		umax, err := strconv.ParseUint(params[1], 10, 32)
		if err != nil {
			return nil
		}
		min, max = int(umin), int(umax)
	}

	if min > max {
		return nil
	}

	return func(value interface{}) (string, bool) {
		str, isStr := value.(string)
		if bytes, ok := value.([]byte); ok {
			str = string(bytes)
			isStr = true
		}
		if !isStr {
			return "", false
		}
		if freeSize {
			return str, true
		}
		l := len(str)
		return str, l >= min && l <= max
	}
}
