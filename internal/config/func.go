package config

import "strings"

// SplitURL without empty sets
func SplitURL(url string) []string {
	trimmed := strings.Trim(url, " /\t\r\n")
	split := strings.Split(trimmed, "/")

	// remove empty set when empty url
	if len(split) == 1 && len(split[0]) == 0 {
		return []string{}
	}
	return split
}
