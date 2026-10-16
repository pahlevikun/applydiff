package cv

import (
	"regexp"
	"sort"
)

var extractNumberRe = regexp.MustCompile(`\d+(?:\.\d+)?\+?%?`)

// ExtractNumbers returns sorted unique numeric tokens found in s (for round-trip checks).
func ExtractNumbers(s string) []string {
	found := extractNumberRe.FindAllString(s, -1)
	if len(found) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	var out []string
	for _, tok := range found {
		if _, ok := seen[tok]; ok {
			continue
		}
		seen[tok] = struct{}{}
		out = append(out, tok)
	}
	sort.Strings(out)
	return out
}
