package filter

import (
	"testing"
	"time"

	"github.com/pahlevikun/applydiff/internal/job"
)

func TestPostedWithinHardFilter(t *testing.T) {
	spec := Spec{
		PostedWithin: &IntRule{Hard: true, Max: 5},
	}
	old := job.Job{
		Title:    "Engineer",
		PostedAt: time.Now().Add(-6 * time.Hour).UTC().Format(time.RFC3339),
	}
	if Evaluate(old, spec).Keep {
		t.Fatal("expected skip for old posting")
	}
	fresh := job.Job{
		Title:    "Engineer",
		PostedAt: time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339),
	}
	if !Evaluate(fresh, spec).Keep {
		t.Fatal("expected keep for fresh posting")
	}
}
