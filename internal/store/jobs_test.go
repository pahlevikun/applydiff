package store

import (
	"context"
	"testing"

	"github.com/pahlevikun/applydiff/internal/job"
)

func TestStoreInsertAndListJob(t *testing.T) {
	ctx := context.Background()
	st, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	j := job.Job{
		Company:  "Acme",
		Title:    "Backend Engineer",
		URL:      "https://example.com/jobs/1",
		Source:   "greenhouse",
		PostedAt: "2026-10-01T12:00:00Z",
		Remote:   "yes",
		Region:   "global",
		Status:   "discovered",
	}
	if err := st.InsertJob(ctx, j); err != nil {
		t.Fatal(err)
	}

	list, err := st.ListJobs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 job, got %d", len(list))
	}
	if list[0].Company != "Acme" || list[0].Title != "Backend Engineer" {
		t.Fatalf("unexpected job: %+v", list[0])
	}
}
