package main

import (
	"git.xdrm.io/go/aicra/driver"
)

func main()                  {}
func Export() driver.Checker { return new(BoolChecker) }

type BoolChecker int

// Match matches the string 'bool'
func (bck BoolChecker) Match(name string) bool {
	return name == "bool"
}

// Check returns true for any type from the @validationTable
func (bck BoolChecker) Check(value interface{}) bool {

	// check if bool
	_, ok := value.(bool)
	return ok

}
