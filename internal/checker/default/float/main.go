package main

import (
	"git.xdrm.io/go/aicra/driver"
)

func main()                  {}
func Export() driver.Checker { return new(FloatChecker) }

type FloatChecker int

// Match matches the string 'int'
func (fck FloatChecker) Match(name string) bool {
	return name == "float"
}

// Check returns true for any type from the @validationTable
func (fck FloatChecker) Check(value interface{}) bool {

	// check if float (default wrapping type)
	_, ok := value.(float64)
	return ok

}
