package store

import (
	"context"
	"testing"

	"github.com/pahlevikun/applydiff/internal/job"
)

func TestRiskChannelMigrationOnExistingRows(t *testing.T) {
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
		ID:       "existing-1",
		Company:  "Acme",
		Title:    "Engineer",
		URL:      "https://acme.com/j/1",
		Source:   "greenhouse",
		PostedAt: "2026-10-01T12:00:00Z",
		Status:   "discovered",
	}
	if err := st.InsertJob(ctx, j); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateApplication(ctx, j.ID); err != nil {
		t.Fatal(err)
	}

	got, err := st.GetJob(ctx, j.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Risk != 0 || got.FlagsJSON != "[]" {
		t.Fatalf("job defaults: risk=%d flags=%q", got.Risk, got.FlagsJSON)
	}
	app, err := st.GetApplication(ctx, j.ID)
	if err != nil {
		t.Fatal(err)
	}
	if app.Channel != "board" || app.Referrer != "" || app.Notes != "" {
		t.Fatalf("application defaults: %+v", app)
	}
}
