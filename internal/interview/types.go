package interview

// Kit is generated prep material for one application (job).
type Kit struct {
	ID            string `json:"id"`
	ApplicationID string `json:"application_id"`
	LoopJSON      string `json:"loop_json"`
	TopicsJSON    string `json:"topics_json"`
	PromptsJSON   string `json:"prompts_json"`
	QuestionsJSON string `json:"questions_json"`
	CreatedAt     string `json:"created_at"`
}

// Story is a STAR entry in the story bank.
type Story struct {
	ID            string `json:"id"`
	ApplicationID string `json:"application_id"`
	Category      string `json:"category"`
	Situation     string `json:"situation"`
	Task          string `json:"task"`
	Action        string `json:"action"`
	Result        string `json:"result"`
	Learning      string `json:"learning"`
	EvidenceJSON  string `json:"evidence_json"`
	CreatedAt     string `json:"created_at"`
}

// MockSession is one mock interview run.
type MockSession struct {
	ID            string `json:"id"`
	ApplicationID string `json:"application_id"`
	Type          string `json:"type"`
	Prompt        string `json:"prompt"`
	StartedAt     string `json:"started_at"`
	FinishedAt    string `json:"finished_at"`
	Status        string `json:"status"`
}

// MockStep is a timed step within a mock session.
type MockStep struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	Idx       int    `json:"idx"`
	Name      string `json:"name"`
	TimeboxS  int    `json:"timebox_s"`
	Output    string `json:"output"`
	Note      string `json:"note"`
	Weak      bool   `json:"weak"`
}

// RubricScore is one dimension score for a session.
type RubricScore struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	Dimension string `json:"dimension"`
	Score     int    `json:"score"`
	Comment   string `json:"comment"`
}

// TopicNote is seeded reference or mindset content.
type TopicNote struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Tier  string `json:"tier"`
}

// Drill is a short practice prompt for a topic.
type Drill struct {
	ID        string `json:"id"`
	TopicSlug string `json:"topic_slug"`
	Prompt    string `json:"prompt"`
	Minutes   int    `json:"minutes"`
}

// ListSummary groups interview artifacts for one job.
type ListSummary struct {
	ApplicationID string        `json:"application_id"`
	Kits          []Kit         `json:"kits"`
	Sessions      []MockSession `json:"sessions"`
	Stories       []Story       `json:"stories"`
}
