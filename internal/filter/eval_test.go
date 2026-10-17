package filter

import (
	"testing"

	"github.com/pahlevikun/applydiff/internal/job"
)

func TestHardFilterSkipsWithReason(t *testing.T) {
	spec := Spec{
		VisaSponsorship: &BoolRule{Hard: true, Require: true},
	}
	j := job.Job{Title: "Engineer", VisaSponsorship: false}
	r := Evaluate(j, spec)
	if r.Keep {
		t.Fatal("expected skip")
	}
	if len(r.Reasons) == 0 || r.Reasons[0] != "skipped: no visa sponsorship" {
		t.Fatalf("reasons: %v", r.Reasons)
	}
}

func TestSoftFilterBoostsScore(t *testing.T) {
	spec := Spec{
		Benefits: &CategoryRule{Hard: false, Values: []string{"4-day week"}, Boost: 10},
	}
	j := job.Job{Benefits: []string{"4-day week", "health insurance"}}
	r := Evaluate(j, spec)
	if !r.Keep {
		t.Fatal("expected keep")
	}
	if r.Score != 10 {
		t.Fatalf("score %d", r.Score)
	}
}

func TestANDAcrossCategoriesORWithin(t *testing.T) {
	spec := Spec{
		Arrangement: &CategoryRule{Hard: true, Values: []string{"remote", "hybrid"}},
		Regions:     &CategoryRule{Hard: true, Values: []string{"SEA", "EU"}},
	}
	jOK := job.Job{Arrangement: "remote", Region: "SEA"}
	if !Evaluate(jOK, spec).Keep {
		t.Fatal("remote+SEA should keep")
	}
	jBad := job.Job{Arrangement: "remote", Region: "NA"}
	if Evaluate(jBad, spec).Keep {
		t.Fatal("remote+NA should skip")
	}
	jHybrid := job.Job{Arrangement: "hybrid", Region: "EU"}
	if !Evaluate(jHybrid, spec).Keep {
		t.Fatal("hybrid+EU should keep")
	}
}
