package interview

import (
	"fmt"
	"strings"

	"github.com/pahlevikun/applydiff/internal/cv"
)

// Story categories for behavioral prep.
var StoryCategories = []string{
	"failure",
	"conflict",
	"disagree-and-commit",
	"ambiguity",
	"deadline",
	"influence",
}

func validCategory(c string) bool {
	for _, cat := range StoryCategories {
		if cat == c {
			return true
		}
	}
	return false
}

// ValidateStory checks category and runs the CV facts gate on numeric tokens.
func ValidateStory(facts *cv.Facts, s Story) error {
	if !validCategory(s.Category) {
		return fmt.Errorf("interview: invalid story category %q", s.Category)
	}
	if strings.TrimSpace(s.Situation) == "" || strings.TrimSpace(s.Action) == "" {
		return fmt.Errorf("interview: situation and action are required")
	}
	if facts == nil {
		return fmt.Errorf("facts: no approved facts loaded")
	}
	text := strings.Join([]string{
		s.Situation, s.Task, s.Action, s.Result, s.Learning, s.EvidenceJSON,
	}, "\n")
	return facts.GateNumberIn(text)
}
