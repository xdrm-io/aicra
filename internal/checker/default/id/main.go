package main

import (
	"git.xdrm.io/go/aicra/driver"
	"math"
)

func main()                  {}
func Export() driver.Checker { return new(IdChecker) }

type IdChecker int

// Match matches the string 'id'
func (ick IdChecker) Match(name string) bool {
	return name == "id"
}

// Check returns true for any type from the @validationTable
func (ick IdChecker) Check(value interface{}) bool {

	// check if float (default wrapping type)
	floatVal, ok := value.(float64)
	if !ok {
		return false
	}

	// check if there is no floating point
	return floatVal == math.Floor(floatVal) && floatVal > 0

}
