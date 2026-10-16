package cv

import (
	"encoding/json"
	"fmt"
	"os"
)

type Document struct {
	Person      string        `json:"person"`
	Contact     Contact       `json:"contact"`
	Summary     Summary       `json:"summary"`
	Skills      []SkillGroup  `json:"skills"`
	Experience  []Experience  `json:"experience"`
	Education   []Education   `json:"education"`
	Achievements []Achievement `json:"achievements"`
}

type Contact struct {
	Name         string `json:"name"`
	Headline     string `json:"headline"`
	Email        string `json:"email"`
	Site         string `json:"site"`
	LinkedIn     string `json:"linkedin"`
	GitHub       string `json:"github"`
	Location     string `json:"location"`
	TimezoneLine string `json:"timezone_line"`
}

type Summary struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type SkillGroup struct {
	ID    string   `json:"id"`
	Label string   `json:"label"`
	Items []string `json:"items"`
}

type Experience struct {
	ID       string   `json:"id"`
	Company  string   `json:"company"`
	Team     string   `json:"team"`
	Title    string   `json:"title"`
	Location string   `json:"location"`
	Start    string   `json:"start"`
	End      *string  `json:"end"`
	Bullets  []Bullet `json:"bullets"`
}

type Bullet struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Education struct {
	ID     string `json:"id"`
	School string `json:"school"`
	Degree string `json:"degree"`
	Year   string `json:"year"`
}

type Achievement struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func LoadFile(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadJSON(data)
}

func LoadJSON(data []byte) (*Document, error) {
	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if err := Validate(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func Validate(doc *Document) error {
	if doc == nil {
		return fmt.Errorf("cv: document is nil")
	}
	if doc.Person == "" {
		return fmt.Errorf("cv: person is required")
	}
	if doc.Summary.ID == "" {
		return fmt.Errorf("cv: summary.id is required")
	}
	for i, sg := range doc.Skills {
		if sg.ID == "" {
			return fmt.Errorf("cv: skills[%d].id is required", i)
		}
	}
	for i, exp := range doc.Experience {
		if exp.ID == "" {
			return fmt.Errorf("cv: experience[%d].id is required", i)
		}
		for j, b := range exp.Bullets {
			if b.ID == "" {
				return fmt.Errorf("cv: experience[%d].bullets[%d].id is required", i, j)
			}
		}
	}
	for i, ed := range doc.Education {
		if ed.ID == "" {
			return fmt.Errorf("cv: education[%d].id is required", i)
		}
	}
	for i, ach := range doc.Achievements {
		if ach.ID == "" {
			return fmt.Errorf("cv: achievements[%d].id is required", i)
		}
	}
	return nil
}

func (doc *Document) Clone() *Document {
	data, err := json.Marshal(doc)
	if err != nil {
		return nil
	}
	out, err := LoadJSON(data)
	if err != nil {
		return nil
	}
	return out
}
