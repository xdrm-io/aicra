package main

import (
	"reflect"
)

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

func Match(name string) bool {
	return name == "int"
}

func Check(value interface{}) bool {

	kind := reflect.TypeOf(value).Kind()

	_, isTypeValid := validationTable[kind]

	return isTypeValid

}
