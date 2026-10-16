package cv

import (
	"testing"
)

func TestFactsGateBlocksUnknownNumber(t *testing.T) {
	facts := &Facts{Facts: map[string]string{
		"abac_impl": "40%",
	}}
	if err := facts.GateNumberIn("Improved latency by 40%"); err != nil {
		t.Fatalf("expected 40%% to pass: %v", err)
	}
	if err := facts.GateNumberIn("Improved latency by 99%"); err == nil {
		t.Fatal("expected 99% to be blocked")
	}
}
