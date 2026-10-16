package diff

import (
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/tailor"
)

func doc() *cv.Document {
	return &cv.Document{
		Person:  "p",
		Summary: cv.Summary{ID: "summary", Text: "One"},
		Skills:  []cv.SkillGroup{{ID: "skills.go", Label: "Go", Items: []string{"Go"}}},
		Experience: []cv.Experience{{
			ID: "exp.1", Company: "C", Title: "E", Start: "2020",
			Bullets: []cv.Bullet{{ID: "exp.1.b1", Text: "Before text"}},
		}},
		Education:    []cv.Education{{ID: "edu.1", School: "S", Degree: "BS", Year: "2020"}},
		Achievements: []cv.Achievement{{ID: "ach.1", Text: "A"}},
	}
}

func TestApplyRevertRoundTrip(t *testing.T) {
	master := doc()
	h := tailor.Hunk{FieldID: "exp.1.b1", Before: "Before text", After: "After text", Status: tailor.StatusPending}
	applied, err := Apply(master, h)
	if err != nil {
		t.Fatal(err)
	}
	text, _ := applied.Field("exp.1.b1")
	if text != "After text" {
		t.Fatalf("apply failed: %q", text)
	}
	reverted, err := Revert(applied, h)
	if err != nil {
		t.Fatal(err)
	}
	text, _ = reverted.Field("exp.1.b1")
	if text != "Before text" {
		t.Fatalf("revert failed: %q", text)
	}
}

func TestApplyAllAcceptedMatchesMasterPlusAccepted(t *testing.T) {
	master := doc()
	hunks := []tailor.Hunk{
		{FieldID: "summary", Before: "One", After: "Two", Status: tailor.StatusAccepted},
		{FieldID: "exp.1.b1", Before: "Before text", After: "Shipped APIs", Status: tailor.StatusPending},
		{FieldID: "exp.1.b1", Before: "Before text", After: "Accepted bullet", Status: tailor.StatusAccepted},
	}
	got, err := ApplyAllAccepted(master, hunks)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := got.Field("summary")
	if s != "Two" {
		t.Fatalf("summary: %q", s)
	}
	b, _ := got.Field("exp.1.b1")
	if b != "Accepted bullet" {
		t.Fatalf("bullet: %q", b)
	}
	masterText, _ := master.Field("exp.1.b1")
	if masterText != "Before text" {
		t.Fatal("master mutated")
	}
}
