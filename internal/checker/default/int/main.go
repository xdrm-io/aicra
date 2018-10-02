package main

import (
	"git.xdrm.io/go/aicra/driver"
	"reflect"
)

func main()                  {}
func Export() driver.Checker { return new(IntChecker) }

var validationTable = map[reflect.Kind]interface{}{
	reflect.Float32: nil,
	reflect.Float64: nil,
	reflect.Int:     nil,
	reflect.Int8:    nil,
	reflect.Int16:   nil,
	reflect.Int32:   nil,
	reflect.Int64:   nil,
	reflect.Uint:    nil,
	reflect.Uint8:   nil,
	reflect.Uint16:  nil,
	reflect.Uint32:  nil,
	reflect.Uint64:  nil,
}

type IntChecker int

// Match matches the string 'int'
func (ick IntChecker) Match(name string) bool {
	return name == "int"
}

// Check returns true for any type from the @validationTable
func (ick IntChecker) Check(value interface{}) bool {

	kind := reflect.TypeOf(value).Kind()

	_, isTypeValid := validationTable[kind]

	return isTypeValid

}
