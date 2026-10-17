package source

import "strings"

// MatchesQueries returns true when queries is empty or any query matches any of the texts.
func MatchesQueries(queries []string, texts ...string) bool {
	if len(queries) == 0 {
		return true
	}
	combined := strings.ToLower(strings.Join(texts, " "))
	for _, q := range queries {
		q = strings.TrimSpace(strings.ToLower(q))
		if q == "" {
			continue
		}
		if strings.Contains(combined, q) {
			return true
		}
	}
	return false
}
