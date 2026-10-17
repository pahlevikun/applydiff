package theme

import (
	"regexp"
	"strings"
)

// CSS is embedded from web/assets/theme.css (single source of Phosphor tokens).
var CSS string

var vars map[string]string

var reVar = regexp.MustCompile(`--([a-z0-9_-]+)\s*:\s*(#[0-9a-fA-F]{3,8})`)

// Init parses CSS custom properties from the canonical theme stylesheet.
func Init(css string) {
	CSS = css
	vars = make(map[string]string)
	for _, m := range reVar.FindAllStringSubmatch(css, -1) {
		if len(m) == 3 {
			vars[m[1]] = m[2]
		}
	}
}

// Color returns a token value or empty if unknown.
func Color(name string) string {
	if vars == nil {
		return ""
	}
	return vars[name]
}

// MustColor returns the token or a visible fallback for TUI when unset.
func MustColor(name, fallback string) string {
	c := Color(name)
	if c == "" {
		return fallback
	}
	return c
}

// All returns a copy of parsed variables.
func All() map[string]string {
	out := make(map[string]string, len(vars))
	for k, v := range vars {
		out[k] = v
	}
	return out
}

// ParseFileContent is used by tests and embed init.
func ParseFileContent(css string) map[string]string {
	out := make(map[string]string)
	for _, m := range reVar.FindAllStringSubmatch(css, -1) {
		if len(m) == 3 {
			out[m[1]] = m[2]
		}
	}
	return out
}

// HasHexOutsideTokens reports whether s contains hex colors other than allowed names.
func HasHexOutsideTokens(s string, allowed map[string]bool) bool {
	hexRe := regexp.MustCompile(`#[0-9a-fA-F]{3,8}`)
	for _, h := range hexRe.FindAllString(s, -1) {
		h = strings.ToLower(h)
		if !allowed[h] {
			return true
		}
	}
	return false
}
