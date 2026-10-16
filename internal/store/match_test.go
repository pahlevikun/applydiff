package store

import (
	"context"
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/match"
)

func TestMatchResultRoundTrip(t *testing.T) {
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
		ID:          "job-match-1",
		Company:     "Acme",
		Title:       "Engineer",
		URL:         "https://boards.greenhouse.io/acme/jobs/1",
		Source:      "greenhouse",
		PostedAt:    "2026-10-01T12:00:00Z",
		Status:      "discovered",
		Description: "- Required: Go experience",
	}
	if err := st.InsertJob(ctx, j); err != nil {
		t.Fatal(err)
	}

	doc := &cv.Document{
		Person:  "p",
		Contact: cv.Contact{Name: "A", Headline: "Eng", Email: "a@b.c"},
		Summary: cv.Summary{ID: "summary", Text: "Go dev"},
		Skills:  []cv.SkillGroup{{ID: "skills.backend", Label: "B", Items: []string{"Go"}}},
		Experience: []cv.Experience{{
			ID: "exp.1", Company: "C", Title: "Eng", Start: "2020-01",
			Bullets: []cv.Bullet{{ID: "exp.1.b1", Text: "x"}},
		}},
		Education:    []cv.Education{{ID: "edu.1", School: "S", Degree: "BS", Year: "2020"}},
		Achievements: []cv.Achievement{{ID: "ach.1", Text: "y"}},
	}
	if err := st.SaveMaster(ctx, doc); err != nil {
		t.Fatal(err)
	}

	res := match.Evaluate(match.Input{Job: j, CV: doc})
	if err := st.SaveMatchResult(ctx, res); err != nil {
		t.Fatal(err)
	}
	loaded, err := st.LoadMatchResult(ctx, j.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Total != res.Total {
		t.Fatalf("total mismatch: %d vs %d", loaded.Total, res.Total)
	}
	if len(loaded.Requirements) != len(res.Requirements) {
		t.Fatalf("requirements len %d vs %d", len(loaded.Requirements), len(res.Requirements))
	}
}
