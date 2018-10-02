package main

import (
	"git.xdrm.io/go/aicra/driver"
	"reflect"
)

func main()                  {}
func Export() driver.Checker { return new(StringChecker) }

type StringChecker int

func (sck StringChecker) Match(name string) bool {
	return name == "string"
}

func (sck StringChecker) Check(value interface{}) bool {

	if value == nil {
		return false
	}

	kind := reflect.TypeOf(value).Kind()

	return kind == reflect.String

}
