package runtime

import (
	"context"
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/store"
	"github.com/pahlevikun/applydiff/internal/tailor"
)

func TestBuildGateUndecidedExit3(t *testing.T) {
	ctx := context.Background()
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	jobID := "job-build-gate"
	if err := st.InsertJob(ctx, job.Job{
		ID: jobID, Company: "Co", Title: "Eng", URL: "https://example.com/z",
		Source: "stub", PostedAt: "2026-10-10T10:00:00Z", Remote: "yes", Region: "eu", Status: "matched",
	}); err != nil {
		t.Fatal(err)
	}
	doc := minimalMasterCV()
	if err := st.SaveMaster(ctx, doc); err != nil {
		t.Fatal(err)
	}
	_, err = st.SaveProposal(ctx, jobID, []tailor.Hunk{{
		FieldID: "summary", Section: "summary", Before: "a", After: "b", Rationale: "r",
	}})
	if err != nil {
		t.Fatal(err)
	}
	app := &App{Store: st, Config: &jobconfig.Config{}}
	_, err = app.BuildPDF(ctx, jobID, true, ActorHuman)
	if err == nil {
		t.Fatal("expected build gate error")
	}
	if ExitCode(err) != ExitNeedsApproval {
		t.Fatalf("exit code %d want %d err=%v", ExitCode(err), ExitNeedsApproval, err)
	}
	var na ErrNeedsApproval
	if e, ok := err.(ErrNeedsApproval); ok {
		na = e
	} else {
		t.Fatalf("expected ErrNeedsApproval, got %T", err)
	}
	if len(na.Undecided) != 1 || na.Undecided[0] != 1 {
		t.Fatalf("undecided indexes %v", na.Undecided)
	}
}

func minimalMasterCV() *cv.Document {
	return &cv.Document{
		Person:  "p",
		Contact: cv.Contact{Name: "A", Headline: "Eng", Email: "a@b.c"},
		Summary: cv.Summary{ID: "summary", Text: "a"},
		Skills:  []cv.SkillGroup{{ID: "skills.backend", Label: "B", Items: []string{"Go"}}},
		Experience: []cv.Experience{{
			ID: "exp.1", Company: "C", Title: "Eng", Start: "2020-01",
			Bullets: []cv.Bullet{{ID: "exp.1.b1", Text: "Built APIs"}},
		}},
		Education:    []cv.Education{{ID: "edu.1", School: "S", Degree: "BS", Year: "2020"}},
		Achievements: []cv.Achievement{{ID: "ach.1", Text: "Award"}},
	}
}
