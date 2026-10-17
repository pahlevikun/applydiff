package interview

// Step is one timed segment in a mock framework.
type Step struct {
	Idx            int      `json:"idx"`
	Name           string   `json:"name"`
	Minutes        int      `json:"minutes"`
	RequiredOutput string   `json:"required_output"`
	Probes         []string `json:"probes"`
}

// Framework defines a mock interview type.
type Framework struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	TotalMinutes  int      `json:"total_minutes"`
	Steps         []Step   `json:"steps"`
	RubricDims    []string `json:"rubric_dimensions"`
}

var frameworks = []Framework{
	{
		ID:           "system-design",
		Name:         "System design",
		TotalMinutes: 45,
		Steps: []Step{
			{Idx: 0, Name: "Clarify requirements", Minutes: 5, RequiredOutput: "scope + functional/non-functional", Probes: []string{"Who are the users?", "What is in scope for v1?"}},
			{Idx: 1, Name: "Estimate scale", Minutes: 5, RequiredOutput: "QPS, storage, bandwidth", Probes: []string{"Reads vs writes?", "Retention period?"}},
			{Idx: 2, Name: "API & data model", Minutes: 5, RequiredOutput: "endpoints + entities", Probes: []string{"Primary entities?", "Idempotency keys?"}},
			{Idx: 3, Name: "High-level architecture", Minutes: 10, RequiredOutput: "diagram + request flow", Probes: []string{"Sync vs async path?", "Failure modes?"}},
			{Idx: 4, Name: "Deep dive 1-2 areas", Minutes: 15, RequiredOutput: "problem -> solution -> tradeoffs", Probes: []string{"Bottleneck?", "What breaks at 10x?"}},
			{Idx: 5, Name: "Wrap up", Minutes: 5, RequiredOutput: "summary + gaps + one question", Probes: []string{"What would you monitor?", "What is your question for us?"}},
		},
		RubricDims: []string{"communication", "problem-solving", "depth", "scalability", "tradeoffs"},
	},
	{
		ID:           "coding",
		Name:         "Coding",
		TotalMinutes: 45,
		Steps: []Step{
			{Idx: 0, Name: "Clarify", Minutes: 5, RequiredOutput: "constraints + edge cases", Probes: []string{"Input size?", "Duplicates allowed?"}},
			{Idx: 1, Name: "Examples", Minutes: 5, RequiredOutput: "worked examples", Probes: []string{"Empty input?", "Single element?"}},
			{Idx: 2, Name: "Brute force", Minutes: 5, RequiredOutput: "naive approach + complexity", Probes: []string{"Can we do better?"}},
			{Idx: 3, Name: "Optimize", Minutes: 10, RequiredOutput: "improved approach + complexity", Probes: []string{"Space/time tradeoff?"}},
			{Idx: 4, Name: "Code", Minutes: 15, RequiredOutput: "working implementation", Probes: []string{"Naming clear?", "Handle edge cases?"}},
			{Idx: 5, Name: "Test", Minutes: 3, RequiredOutput: "test cases walked through", Probes: []string{"Regression case?"}},
			{Idx: 6, Name: "Complexity", Minutes: 2, RequiredOutput: "final time and space complexity", Probes: []string{"Amortized cost?"}},
		},
		RubricDims: []string{"communication", "problem-solving", "correctness", "code-quality", "complexity-analysis"},
	},
	{
		ID:           "behavioral",
		Name:         "Behavioral",
		TotalMinutes: 30,
		Steps: []Step{
			{Idx: 0, Name: "Pick story", Minutes: 3, RequiredOutput: "category + story outline", Probes: []string{"Which STAR category fits?"}},
			{Idx: 1, Name: "Situation & task", Minutes: 5, RequiredOutput: "context and goal", Probes: []string{"Stakes clear?"}},
			{Idx: 2, Name: "Action", Minutes: 15, RequiredOutput: "what you did (majority of time)", Probes: []string{"Specific actions?", "Your role vs team?"}},
			{Idx: 3, Name: "Result & learning", Minutes: 7, RequiredOutput: "outcome + takeaway", Probes: []string{"Measurable result?", "What would you do differently?"}},
		},
		RubricDims: []string{"communication", "structure", "impact", "self-awareness", "collaboration"},
	},
	{
		ID:           "business-case",
		Name:         "Business case",
		TotalMinutes: 45,
		Steps: []Step{
			{Idx: 0, Name: "Clarify objective", Minutes: 5, RequiredOutput: "goal and success metric", Probes: []string{"Who is the client?", "Time horizon?"}},
			{Idx: 1, Name: "Issue tree", Minutes: 10, RequiredOutput: "structured breakdown", Probes: []string{"MECE branches?"}},
			{Idx: 2, Name: "Analysis", Minutes: 15, RequiredOutput: "numbers and assumptions", Probes: []string{"Sensitivity?", "Data gaps?"}},
			{Idx: 3, Name: "Recommendation", Minutes: 10, RequiredOutput: "clear recommendation", Probes: []string{"Why this over alternatives?"}},
			{Idx: 4, Name: "Risks", Minutes: 5, RequiredOutput: "risks and mitigations", Probes: []string{"What could invalidate the plan?"}},
		},
		RubricDims: []string{"communication", "problem-solving", "analytical-rigor", "synthesis", "judgment"},
	},
	{
		ID:           "product-case",
		Name:         "Product case",
		TotalMinutes: 45,
		Steps: []Step{
			{Idx: 0, Name: "Problem", Minutes: 5, RequiredOutput: "problem statement", Probes: []string{"Who hurts most?"}},
			{Idx: 1, Name: "User", Minutes: 5, RequiredOutput: "persona and jobs-to-be-done", Probes: []string{"Primary user segment?"}},
			{Idx: 2, Name: "Metrics", Minutes: 5, RequiredOutput: "north-star and guardrails", Probes: []string{"Leading vs lagging?"}},
			{Idx: 3, Name: "Options", Minutes: 10, RequiredOutput: "solution options", Probes: []string{"Build vs buy?"}},
			{Idx: 4, Name: "Prioritize", Minutes: 10, RequiredOutput: "RICE or equivalent ranking", Probes: []string{"Why top item first?"}},
			{Idx: 5, Name: "Tradeoffs", Minutes: 10, RequiredOutput: "tradeoffs and rollout", Probes: []string{"What do we cut?"}},
		},
		RubricDims: []string{"communication", "problem-solving", "product-sense", "prioritization", "tradeoffs"},
	},
}

// FrameworkIDs are valid mock --type values.
var FrameworkIDs = []string{"system-design", "coding", "behavioral", "business-case", "product-case"}

// FrameworkByID returns a framework or nil.
func FrameworkByID(id string) *Framework {
	switch id {
	case "case":
		id = "business-case"
	case "product":
		id = "product-case"
	}
	for i := range frameworks {
		if frameworks[i].ID == id {
			f := frameworks[i]
			return &f
		}
	}
	return nil
}

// AllFrameworks returns the built-in framework registry.
func AllFrameworks() []Framework {
	out := make([]Framework, len(frameworks))
	copy(out, frameworks)
	return out
}
