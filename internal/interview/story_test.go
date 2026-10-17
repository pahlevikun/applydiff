package interview

import (
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
)

func TestValidateStoryFactsGate(t *testing.T) {
	facts := &cv.Facts{Facts: map[string]string{
		"latency_win": "Reduced p99 latency by 40%",
	}}
	grounded := Story{
		Category:  "failure",
		Situation: "Service degraded",
		Task:      "Restore stability",
		Action:    "Rolled back and fixed root cause",
		Result:    "Reduced p99 latency by 40%",
		Learning:  "Better canaries",
		EvidenceJSON: `["latency_win"]`,
	}
	if err := ValidateStory(facts, grounded); err != nil {
		t.Fatalf("grounded story should pass: %v", err)
	}
	invented := grounded
	invented.Result = "Improved throughput by 99%"
	if err := ValidateStory(facts, invented); err == nil {
		t.Fatal("expected invented metric to be rejected")
	}
}
