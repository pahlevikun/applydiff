package store

import (
	"context"
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/tailor"
)

func TestProposalHunkAcceptRejectBranchCV(t *testing.T) {
	ctx := context.Background()
	st, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	master := &cv.Document{
		Person:  "person",
		Summary: cv.Summary{ID: "summary", Text: "Master summary"},
		Skills: []cv.SkillGroup{
			{ID: "skills.go", Label: "Go", Items: []string{"Go"}},
		},
		Experience: []cv.Experience{{
			ID: "exp.1", Company: "Co", Title: "Eng", Start: "2020-01",
			Bullets: []cv.Bullet{{ID: "exp.1.b1", Text: "Built APIs"}},
		}},
		Education:    []cv.Education{{ID: "edu.1", School: "U", Degree: "BS", Year: "2020"}},
		Achievements: []cv.Achievement{{ID: "ach.1", Text: "Award"}},
	}
	if err := st.SaveMaster(ctx, master); err != nil {
		t.Fatal(err)
	}

	hunks := []tailor.Hunk{
		{FieldID: "exp.1.b1", Section: "experience", Before: "Built APIs", After: "Built Go APIs", Status: tailor.StatusPending},
		{FieldID: "summary", Section: "summary", Before: "Master summary", After: "Rejected summary", Status: tailor.StatusPending},
	}
	prop, err := st.SaveProposal(ctx, "job-1", hunks)
	if err != nil {
		t.Fatal(err)
	}
	if len(prop.Hunks) != 2 {
		t.Fatalf("expected 2 hunks, got %d", len(prop.Hunks))
	}

	if err := st.SetHunkStatus(ctx, prop.Hunks[0].ID, tailor.StatusAccepted); err != nil {
		t.Fatal(err)
	}
	if err := st.SetHunkStatus(ctx, prop.Hunks[1].ID, tailor.StatusRejected); err != nil {
		t.Fatal(err)
	}

	branch, err := st.LoadBranch(ctx, "job-1")
	if err != nil {
		t.Fatal(err)
	}
	text, _ := branch.Field("exp.1.b1")
	if text != "Built Go APIs" {
		t.Fatalf("expected accepted hunk on branch, got %q", text)
	}
	sum, _ := branch.Field("summary")
	if sum != "Master summary" {
		t.Fatalf("rejected hunk should not apply, got %q", sum)
	}
}
