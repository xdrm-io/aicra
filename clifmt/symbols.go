package clifmt

import (
	"fmt"
)

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
