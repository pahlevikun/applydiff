package tailor

import (
	"encoding/json"
	"fmt"
	"strings"
)

type responseEnvelope struct {
	Hunks []RawHunk `json:"hunks"`
}

// ParseHunksJSON extracts hunks from an LLM JSON response (with optional markdown fences).
func ParseHunksJSON(text string) ([]RawHunk, error) {
	text = strings.TrimSpace(text)
	text = stripCodeFence(text)
	var env responseEnvelope
	if err := json.Unmarshal([]byte(text), &env); err != nil {
		return nil, fmt.Errorf("tailor: parse hunks json: %w", err)
	}
	if len(env.Hunks) == 0 {
		return nil, fmt.Errorf("tailor: no hunks in response")
	}
	return env.Hunks, nil
}

func stripCodeFence(s string) string {
	if !strings.HasPrefix(s, "```") {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) < 2 {
		return s
	}
	start := 1
	end := len(lines)
	if strings.TrimSpace(lines[len(lines)-1]) == "```" {
		end = len(lines) - 1
	}
	return strings.Join(lines[start:end], "\n")
}
