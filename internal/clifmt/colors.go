package clifmt

import (
	"fmt"
)

// Color returns a bash-formatted string representing
// the string @text with the color code @color and in bold
// if @bold (1 optional argument) is set to true
func Color(color byte, text string, bold ...bool) string {
	b := "0"
	if len(bold) > 0 && bold[0] {
		b = "1"
	}
	return fmt.Sprintf("\033[%s;%dm%s\033[0m", b, color, text)
}
