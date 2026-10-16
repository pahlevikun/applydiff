package cv

import (
	"encoding/json"
	"testing"
)

func TestLoadValidCV(t *testing.T) {
	doc := sampleDoc()
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Summary.ID != "summary" {
		t.Fatalf("unexpected summary id: %s", loaded.Summary.ID)
	}
}

func TestRejectMissingBulletID(t *testing.T) {
	doc := sampleDoc()
	doc.Experience[0].Bullets[0].ID = ""
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	_, err = LoadJSON(data)
	if err == nil {
		t.Fatal("expected validation error for missing bullet id")
	}
}

func sampleDoc() *Document {
	end := "2022-12"
	return &Document{
		Person: "test-person",
		Contact: Contact{
			Name:     "Test User",
			Headline: "Engineer",
			Email:    "test@example.com",
		},
		Summary: Summary{ID: "summary", Text: "Summary text"},
		Skills: []SkillGroup{
			{ID: "skills.backend", Label: "Backend", Items: []string{"Go"}},
		},
		Experience: []Experience{
			{
				ID: "exp.1", Company: "Co", Title: "Eng", Start: "2020-01", End: &end,
				Bullets: []Bullet{{ID: "exp.1.b1", Text: "Did things"}},
			},
		},
		Education: []Education{
			{ID: "edu.1", School: "Uni", Degree: "BS", Year: "2020"},
		},
		Achievements: []Achievement{{ID: "ach.1", Text: "Award"}},
	}
}
