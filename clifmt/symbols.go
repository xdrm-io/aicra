package clifmt

import (
	"fmt"
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
	fmt.Printf("\n%s (%d) %s %s\n", Color(33, ">>", false), title_index, s, Color(33, "<<", false))

}

func Align(s string) {
	fmt.Printf("%s", s)
	for i := len(s); i < align_offset; i++ {
		fmt.Printf(" ")
	}
}
