package job

import (
	"html"
	"regexp"
	"strings"
)

var htmlTagRe = regexp.MustCompile(`(?i)<[^>]+>`)
var parenRe = regexp.MustCompile(`[()]`)

var legalSuffixes = map[string]struct{}{
	"inc": {}, "ltd": {}, "gmbh": {}, "llc": {}, "pte": {},
}

// RoleKey identifies the same role across boards (company + normalized title).
// DedupeKey (per URL) remains the upsert identity; role_key is the cross-board grouping key.
func RoleKey(company, title string) string {
	c := normalizeCompany(company)
	t := normalizeTitle(title)
	if c == "" || t == "" {
		return ""
	}
	return c + "\x00" + t
}

func normalizeCompany(company string) string {
	s := strings.ToLower(strings.TrimSpace(company))
	s = strings.Join(strings.Fields(s), " ")
	if s == "" {
		return ""
	}
	parts := strings.Fields(s)
	if len(parts) > 1 {
		last := parts[len(parts)-1]
		if _, ok := legalSuffixes[last]; ok {
			parts = parts[:len(parts)-1]
			s = strings.Join(parts, " ")
		}
	}
	return s
}

func normalizeTitle(title string) string {
	s := strings.TrimSpace(title)
	if s == "" {
		return ""
	}
	s = htmlTagRe.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	s = strings.ToLower(s)
	s = parenRe.ReplaceAllString(s, " ")
	s = strings.Join(strings.Fields(s), " ")
	if s == "" {
		return ""
	}
	s = stripOneTrailingQualifier(s)
	return strings.TrimSpace(s)
}

func stripOneTrailingQualifier(s string) string {
	seps := []string{",", "-", "|", "@"}
	best := -1
	for _, sep := range seps {
		idx := strings.LastIndex(s, sep)
		if idx > 0 && (best < 0 || idx > best) {
			best = idx
		}
	}
	if best > 0 {
		left := strings.TrimSpace(s[:best])
		if left != "" {
			return left
		}
	}
	return s
}
