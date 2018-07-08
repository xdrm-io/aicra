package clifmt

import (
	"fmt"
	"strings"
)

var title_index = 0
var align_offset = 30

func Warn(s ...string) string {
	if len(s) == 0 {
		return Color(31, "/!\\")
	}

	return fmt.Sprintf("%s  %s", Warn(), s[0])
}
func Info(s ...string) string {
	if len(s) == 0 {
		return Color(34, "(!)")
	}

	return fmt.Sprintf("%s  %s", Info(), s[0])
}

func Title(s string) {
	title_index++
	fmt.Printf("\n%s |%d| %s %s\n", Color(33, ">>", false), title_index, s, Color(33, "<<", false))

}

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
	for i := size; i < align_offset; i++ {
		fmt.Printf(" ")
	}
}
