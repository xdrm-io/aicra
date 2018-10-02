package main

import (
	"git.xdrm.io/go/aicra/driver"
	"math"
)

func main()                  {}
func Export() driver.Checker { return new(IntChecker) }

type IntChecker int

// Match matches the string 'int'
func (ick IntChecker) Match(name string) bool {
	return name == "int"
}

// Check returns true for any type from the @validationTable
func (ick IntChecker) Check(value interface{}) bool {

	// check if float (default wrapping type)
	floatVal, ok := value.(float64)
	if !ok {
		return false
	}

	// check if there is no floating point
	return floatVal == math.Floor(floatVal)

}
