package risk

import "encoding/json"

const (
	SeverityHard = "hard"
	SeveritySoft = "soft"
)

// Flag is a single risk signal with human-readable context.
type Flag struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// Assessment is the deterministic risk output for a job posting.
type Assessment struct {
	Score int    `json:"score"`
	Flags []Flag `json:"flags"`
}

// Context carries optional peer data for cross-job rules.
type Context struct {
	PeerDescriptionHashes map[string]bool
}

func (a Assessment) FlagsJSON() string {
	raw, err := json.Marshal(a.Flags)
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func ParseFlagsJSON(raw string) []Flag {
	if raw == "" || raw == "[]" {
		return nil
	}
	var flags []Flag
	if err := json.Unmarshal([]byte(raw), &flags); err != nil {
		return nil
	}
	return flags
}
