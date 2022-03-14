package reqdata

import (
	"strings"
)

// Query represents an http query (url or body)
type Query map[string][]string

// Parse url encoded data from an url or a raw body
// Simplified version of net/url.parseQuery()
func (q Query) Parse(query string) error {
	for query != "" {
		key := query
		if i := strings.IndexRune(key, '&'); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.IndexRune(key, '='); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		q[key] = append(q[key], value)
	}
	return nil
}
