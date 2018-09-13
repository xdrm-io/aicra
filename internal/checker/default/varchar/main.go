package main

import (
	"regexp"
	"strconv"
)

var min *uint64
var max *uint64

// Match filters the parameter type format "varchar(min, max)"
func Match(name string) bool {

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
	min = &minVal

	/* (4) Extract max */
	maxVal, err := strconv.ParseUint(matches[2], 10, 64)
	if err != nil {
		return false
	}

	/* (5) Check that min <= max */
	if maxVal < minVal {
		panic("varchar(x, y) ; constraint violation : x <= y")
	}
	max = &maxVal

	return true

}

// Check whether the given value fulfills the condition (min, max)
func Check(value interface{}) bool {

	/* (1) Check if string */
	strval, ok := value.(string)
	if !ok {
		return false
	}

	/* (2) Check if sizes set */
	if min == nil || max == nil {
		return false
	}

	/* (3) Check min */
	if uint64(len(strval)) < *min {
		return false
	}

	/* (4) Check max */
	if uint64(len(strval)) > *max {
		return false
	}

	return true

}
