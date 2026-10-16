package tailor

import (
	"regexp"
	"strings"
)

var jdTermRe = regexp.MustCompile(`[a-z][a-z0-9+.#-]{2,}`)

// JDKeywords extracts lowercase keyword-like tokens from a job description.
func JDKeywords(description string) map[string]struct{} {
	desc := strings.ToLower(description)
	tokens := jdTermRe.FindAllString(desc, -1)
	out := make(map[string]struct{})
	for _, tok := range tokens {
		if len(tok) < 4 {
			continue
		}
		out[tok] = struct{}{}
	}
	return out
}

// AddsJDKeyword reports whether after adds a JD term not present in before.
func AddsJDKeyword(before, after string, jdTerms map[string]struct{}) bool {
	if len(jdTerms) == 0 {
		return false
	}
	b := tokenSet(before)
	a := tokenSet(after)
	for term := range jdTerms {
		if a[term] && !b[term] {
			return true
		}
	}
	return false
}

func tokenSet(s string) map[string]bool {
	s = strings.ToLower(s)
	toks := jdTermRe.FindAllString(s, -1)
	out := make(map[string]bool)
	for _, t := range toks {
		out[t] = true
	}
	return out
}
