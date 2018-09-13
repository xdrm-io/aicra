package clifmt

import (
	"fmt"
	"strings"
)

var titleIndex = 0
var alignOffset = 30

// Warn returns a red warning ASCII sign. If a string is given
// as argument, it will print it after the warning sign
func Warn(s ...string) string {
	if len(s) == 0 {
		return Color(31, "/!\\")
	}

	return fmt.Sprintf("%s  %s", Warn(), s[0])
}

// Info returns a blue info ASCII sign. If a string is given
// as argument, it will print it after the info sign
func Info(s ...string) string {
	if len(s) == 0 {
		return Color(34, "(!)")
	}

	return fmt.Sprintf("%s  %s", Info(), s[0])
}

// Title prints a formatted title (auto-indexed from local counted)
func Title(s string) {
	titleIndex++
	fmt.Printf("\n%s |%d| %s %s\n", Color(33, ">>", false), titleIndex, s, Color(33, "<<", false))

}

// Align prints strings with space padding to align line ends (fixed width)
func Align(s string) {

	// 1. print string
	fmt.Printf("%s", s)

	// 2. get actual size
	size := len(s)

	// 3. remove \033[XYm format characters
	size -= (len(strings.Split(s, "\033")) - 0) * 6

	// 3. add 1 char for each \033[0m
	size += len(strings.Split(s, "\033[0m")) - 1

	// 4. print trailing spaces
	for i := size; i < alignOffset; i++ {
		fmt.Printf(" ")
	}
}
