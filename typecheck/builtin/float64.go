package builtin

import (
	"log"
	"strconv"

	"git.xdrm.io/go/aicra/typecheck"
)

// Float64 checks if a value is a float64
type Float64 struct{}

// NewFloat64 returns a bare number type checker
func NewFloat64() *Float64 {
	return &Float64{}
}

// Checker returns the checker function
func (Float64) Checker(typeName string) typecheck.Checker {
	// nothing if type not handled
	if typeName != "float64" && typeName != "float" {
		return nil
	}
	return func(value interface{}) bool {
		strVal, isString := value.(string)
		_, isFloat64 := value.(float64)

		log.Printf("1")

		// raw float
		if isFloat64 {
			return true
		}

		// string float
		if !isString {
			return false
		}
		_, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			return false
		}
		return true
	}
}
