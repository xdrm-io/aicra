package main

import (
	"git.xdrm.io/go/aicra/driver"
	"regexp"
	"strconv"
)

func main()                  {}
func Export() driver.Checker { return new(VarcharChecker) }

type VarcharChecker struct {
	min *uint64
	max *uint64
}

// Match filters the parameter type format "varchar(min, max)"
func (vck *VarcharChecker) Match(name string) bool {

	vck.min = nil
	vck.max = nil

	/* (1) Create regexp */
	re, err := regexp.Compile(`^varchar\((\d+), ?(\d+)\)$`)
	if err != nil {
		panic(err)
	}

	/* (2) Check if matches */
	matches := re.FindStringSubmatch(name)
	if matches == nil || len(matches) < 3 {
		return false
	}

	/* (3) Extract min */
	minVal, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return false
	}
	vck.min = &minVal

	/* (4) Extract max */
	maxVal, err := strconv.ParseUint(matches[2], 10, 64)
	if err != nil {
		return false
	}

	/* (5) Check that min <= max */
	if maxVal < minVal {
		panic("varchar(x, y) ; constraint violation : x <= y")
	}
	vck.max = &maxVal

	return true

}

// Check whether the given value fulfills the condition (min, max)
func (vck *VarcharChecker) Check(value interface{}) bool {

	/* (1) Check if string */
	strval, ok := value.(string)
	if !ok {
		return false
	}

	/* (2) Check if sizes set */
	if vck.min == nil || vck.max == nil {
		return false
	}

	length := uint64(len(strval))

	/* (3) Check min */
	if length < *vck.min {
		return false
	}

	/* (4) Check max */
	if length > *vck.max {
		return false
	}

	return true

}
