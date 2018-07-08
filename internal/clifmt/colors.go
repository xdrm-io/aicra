package clifmt

import (
	"fmt"
)

func Color(color byte, s string, bold ...bool) string {
	b := "0"
	if len(bold) > 0 && bold[0] {
		b = "1"
	}
	return fmt.Sprintf("\033[%s;%dm%s\033[0m", b, color, s)
}
