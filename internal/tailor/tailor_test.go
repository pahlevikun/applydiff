package tailor

import (
	"context"
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/llm"
)

func sampleMaster() *cv.Document {
	return &cv.Document{
		Person:  "p",
		Contact: cv.Contact{Name: "A", Headline: "Eng", Email: "a@b.c"},
		Summary: cv.Summary{ID: "summary", Text: "Backend engineer with Go"},
		Skills: []cv.SkillGroup{
			{ID: "skills.backend", Label: "Backend", Items: []string{"Go", "Rails"}},
		},
		Experience: []cv.Experience{{
			ID: "exp.1", Company: "Co", Title: "Eng", Start: "2020-01",
			Bullets: []cv.Bullet{{ID: "exp.1.b1", Text: "Built APIs with Go"}},
		}},
		Education:    []cv.Education{{ID: "edu.1", School: "S", Degree: "BS", Year: "2020"}},
		Achievements: []cv.Achievement{{ID: "ach.1", Text: "Award"}},
	}
}

func TestTailorProducesHunks(t *testing.T) {
	canned := `{"hunks":[{"field_id":"exp.1.b1","section":"experience","before":"Built APIs with Go","after":"Architected distributed APIs with Go and Kafka","rationale":"mirror posting","evidence":[]}]}`
	eng := &Engine{Runner: llm.FakeRunner{Response: canned}}
	j := job.Job{Company: "Acme", Title: "Backend", Description: "We need distributed systems and Kafka experience."}
	hunks, err := eng.Generate(context.Background(), j, sampleMaster(), &cv.Facts{Facts: map[string]string{"x": "40%"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}
	if hunks[0].Status != StatusPending {
		t.Fatalf("expected pending, got %s", hunks[0].Status)
	}
}

func TestFactsGateBlocksFabricatedNumber(t *testing.T) {
	raw := []RawHunk{{
		FieldID: "exp.1.b1",
		Section: "experience",
		Before:  "Built APIs with Go",
		After:   "Improved latency by 99%",
	}}
	facts := &cv.Facts{Facts: map[string]string{"abac": "40%"}}
	hunks, err := ValidateHunks(raw, sampleMaster(), facts, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(hunks) != 1 || hunks[0].Status != StatusBlocked {
		t.Fatalf("expected blocked hunk, got %+v", hunks)
	}
}

func TestATSKeywordTag(t *testing.T) {
	raw := []RawHunk{{
		FieldID: "skills.backend",
		Section: "skills",
		Before:  "Go, Rails",
		After:   "Go, Rails, kubernetes orchestration",
	}}
	jd := "Required: kubernetes experience for distributed systems"
	hunks, err := ValidateHunks(raw, sampleMaster(), nil, jd)
	if err != nil {
		t.Fatal(err)
	}
	if len(hunks) != 1 {
		t.Fatal("expected one hunk")
	}
	found := false
	for _, tag := range hunks[0].Tags {
		if tag == TagATSKeyword {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected ats_keyword tag, tags=%v", hunks[0].Tags)
	}
}
