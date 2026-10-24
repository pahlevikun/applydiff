package store

import (
	"context"
	"testing"

	"github.com/pahlevikun/applydiff/internal/job"
)

func TestUpsertPreservesFirstSeenAndStatusOnRescan(t *testing.T) {
	ctx := context.Background()
	st, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	j1 := job.Job{
		Company:  "Acme",
		Title:    "Backend Engineer",
		URL:      "https://example.com/jobs/1",
		Source:   "alpha",
		PostedAt: "2026-10-01T12:00:00Z",
		Status:   "discovered",
	}
	if err := st.UpsertJob(ctx, j1, job.DedupeKey(j1)); err != nil {
		t.Fatal(err)
	}
	list, err := st.ListJobs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 job, got %d", len(list))
	}
	firstSeen := list[0].FirstSeen
	if firstSeen == "" {
		t.Fatal("expected first_seen set on insert")
	}

	if err := st.UpdateJobStatus(ctx, list[0].ID, "matched"); err != nil {
		t.Fatal(err)
	}

	j2 := j1
	j2.PostedAt = "2026-10-24T12:00:00Z"
	j2.Status = "discovered"
	if err := st.UpsertJob(ctx, j2, job.DedupeKey(j2)); err != nil {
		t.Fatal(err)
	}
	list, err = st.ListJobs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if list[0].FirstSeen != firstSeen {
		t.Fatalf("first_seen changed: was %q now %q", firstSeen, list[0].FirstSeen)
	}
	if list[0].Status != "matched" {
		t.Fatalf("status clobbered: %q", list[0].Status)
	}
	if list[0].PostedAt != j2.PostedAt {
		t.Fatalf("posted_at should update: %q", list[0].PostedAt)
	}
}

func TestRoleKeyMergeThreeSourcesOneRow(t *testing.T) {
	ctx := context.Background()
	st, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	base := job.Job{
		Company: "Acme Inc",
		Title:   "Senior Backend Engineer",
		Source:  "one",
		Origins: []string{"one"},
		Status:  "discovered",
	}
	j1 := base
	j1.URL = "https://boards.example.com/acme/1"
	j2 := base
	j2.URL = "https://jobicy.com/jobs/acme-senior"
	j2.Source = "two"
	j2.Origins = []string{"two"}
	j3 := base
	j3.URL = "https://remotive.com/job/acme-backend"
	j3.Source = "three"
	j3.Origins = []string{"three"}

	merged := job.MergeByRoleKey([]job.Job{j1, j2, j3})
	if len(merged) != 1 {
		t.Fatalf("expected 1 merged job, got %d", len(merged))
	}
	if len(merged[0].Origins) < 3 {
		t.Fatalf("expected 3 origins, got %#v", merged[0].Origins)
	}
	if err := st.UpsertJob(ctx, merged[0], job.DedupeKey(merged[0])); err != nil {
		t.Fatal(err)
	}
	n, err := st.CountJobs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 row in db, got %d", n)
	}
}
