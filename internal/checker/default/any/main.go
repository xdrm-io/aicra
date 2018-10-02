package main

import (
	"git.xdrm.io/go/aicra/driver"
)

func main()                  {}
func Export() driver.Checker { return new(AnyChecker) }

type AnyChecker int

// Match matches the string 'any'
func (ack AnyChecker) Match(name string) bool {
	return name == "any"
}

// Check always returns true
func (ack AnyChecker) Check(value interface{}) bool {
	return true
}
