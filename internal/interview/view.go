package interview

// Workspace is the interview prep surface for one job application.
type Workspace struct {
	ApplicationID   string   `json:"application_id"`
	Kit             *Kit     `json:"kit,omitempty"`
	Loop            []string `json:"loop"`
	Topics          []string `json:"topics"`
	Questions       []string `json:"questions"`
	Stories         []Story  `json:"stories"`
	Categories      []string `json:"categories"`
	EmptyCategories []string `json:"empty_categories"`
	MockType        string   `json:"mock_type"`
	Framework       *Framework `json:"framework,omitempty"`
	LatestSession   *MockSession `json:"latest_session,omitempty"`
	LatestSteps     []MockStep   `json:"latest_steps,omitempty"`
	LatestRubric    []RubricScore `json:"latest_rubric,omitempty"`
}

// EmptyStoryCategories returns categories with no stories.
func EmptyStoryCategories(stories []Story) []string {
	have := map[string]bool{}
	for _, s := range stories {
		have[s.Category] = true
	}
	var empty []string
	for _, cat := range StoryCategories {
		if !have[cat] {
			empty = append(empty, cat)
		}
	}
	return empty
}
