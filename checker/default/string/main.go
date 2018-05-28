package main

import (
	"reflect"
)

func Match(name string) bool {
	return name == "string"
}

func Check(value interface{}) bool {

	kind := reflect.TypeOf(value).Kind()

	return kind == reflect.String

}
