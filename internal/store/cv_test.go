package store

import (
	"context"
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
)

func TestCVMasterAndBranch(t *testing.T) {
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
		Experience: []cv.Experience{
			{
				ID: "exp.1", Company: "Co", Title: "Eng", Start: "2020-01",
				Bullets: []cv.Bullet{{ID: "exp.1.b1", Text: "Built APIs"}},
			},
		},
		Education:    []cv.Education{{ID: "edu.1", School: "U", Degree: "BS", Year: "2020"}},
		Achievements: []cv.Achievement{{ID: "ach.1", Text: "Award"}},
	}
	if err := st.SaveMaster(ctx, master); err != nil {
		t.Fatal(err)
	}
	loaded, err := st.LoadMaster(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Summary.Text != "Master summary" {
		t.Fatalf("master round-trip failed: %+v", loaded.Summary)
	}

	if err := st.CreateBranch(ctx, "job-1"); err != nil {
		t.Fatal(err)
	}
	hunk := cv.Hunk{
		FieldID: "exp.1.b1",
		Before:  "Built APIs",
		After:   "Built Go APIs",
	}
	if err := st.ApplyHunk(ctx, "job-1", hunk); err != nil {
		t.Fatal(err)
	}
	branch, err := st.LoadBranch(ctx, "job-1")
	if err != nil {
		t.Fatal(err)
	}
	text, ok := branch.Field("exp.1.b1")
	if !ok || text != "Built Go APIs" {
		t.Fatalf("expected hunk applied, got %q ok=%v", text, ok)
	}
	masterAgain, err := st.LoadMaster(ctx)
	if err != nil {
		t.Fatal(err)
	}
	masterText, _ := masterAgain.Field("exp.1.b1")
	if masterText != "Built APIs" {
		t.Fatalf("master should be unchanged, got %q", masterText)
	}
}
