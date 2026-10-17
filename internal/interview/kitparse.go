package interview

import (
	"encoding/json"
	"strings"
)

// ParseStringListJSON decodes a JSON array of strings from kit fields.
func ParseStringListJSON(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}
