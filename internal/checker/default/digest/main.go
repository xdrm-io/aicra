package main

import (
	"git.xdrm.io/go/aicra/driver"
	"strconv"
	"strings"
)

func main()                  {}
func Export() driver.Checker { return new(DigestChecker) }

type DigestChecker struct {
	Length *uint64
}

// Match filters the parameter type format "varchar(min, max)"
func (dck *DigestChecker) Match(name string) bool {

	dck.Length = nil

	/* (1) Check prefix/suffix */
	if len(name) < len("digest(x)") || !strings.HasPrefix(name, "digest(") || name[len(name)-1] != ')' {
		return false
	}

	/* (2) Extract length */
	lengthStr := name[len("digest(") : len(name)-1]

	length, err := strconv.ParseUint(lengthStr, 10, 64)
	if err != nil {
		return false
	}
	dck.Length = &length

	return true

}

// Check whether the given value fulfills the condition (min, max)
func (dck *DigestChecker) Check(value interface{}) bool {

	/* (1) Check if sizes set */
	if dck.Length == nil {
		return false
	}

	/* (2) Check if string */
	strval, ok := value.(string)
	if !ok {
		return false
	}

	length := uint64(len(strval))

	/* (3) Check length */
	if length != *dck.Length {
		return false
	}

	/* (4) Check character set */
	for _, char := range strval {
		if !isValidCharacter(char) {
			return false
		}
	}

	return true

}

func isValidCharacter(char rune) bool {
	return (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')
}
