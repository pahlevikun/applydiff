package interview

import (
	"testing"
)

func TestFrameworkMinutesSumToTotal(t *testing.T) {
	for _, fw := range AllFrameworks() {
		sum := 0
		for _, st := range fw.Steps {
			sum += st.Minutes
		}
		if sum != fw.TotalMinutes {
			t.Fatalf("%s: step minutes sum %d != total %d", fw.ID, sum, fw.TotalMinutes)
		}
	}
}

func TestSystemDesignStepsMatchSpec(t *testing.T) {
	fw := FrameworkByID("system-design")
	if fw == nil {
		t.Fatal("missing system-design framework")
	}
	if fw.TotalMinutes != 45 {
		t.Fatalf("total minutes %d", fw.TotalMinutes)
	}
	want := []struct {
		name   string
		output string
	}{
		{"Clarify requirements", "scope + functional/non-functional"},
		{"Estimate scale", "QPS, storage, bandwidth"},
		{"API & data model", "endpoints + entities"},
		{"High-level architecture", "diagram + request flow"},
		{"Deep dive 1-2 areas", "problem -> solution -> tradeoffs"},
		{"Wrap up", "summary + gaps + one question"},
	}
	if len(fw.Steps) != len(want) {
		t.Fatalf("steps %d want %d", len(fw.Steps), len(want))
	}
	for i, w := range want {
		st := fw.Steps[i]
		if st.Idx != i {
			t.Fatalf("step %d idx %d", i, st.Idx)
		}
		if st.Name != w.name {
			t.Fatalf("step %d name %q want %q", i, st.Name, w.name)
		}
		if st.RequiredOutput != w.output {
			t.Fatalf("step %d output %q want %q", i, st.RequiredOutput, w.output)
		}
	}
}
