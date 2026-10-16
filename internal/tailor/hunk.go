package tailor

const (
	StatusPending  = "pending"
	StatusAccepted = "accepted"
	StatusRejected = "rejected"
	StatusBlocked  = "blocked"
)

const TagATSKeyword = "ats_keyword"

// Hunk is one proposed tailoring edit on a job branch.
type Hunk struct {
	ID        string   `json:"id"`
	FieldID   string   `json:"field_id"`
	Section   string   `json:"section"`
	Before    string   `json:"before"`
	After     string   `json:"after"`
	Rationale string   `json:"rationale"`
	Evidence  []string `json:"evidence"`
	Tags      []string `json:"tags,omitempty"`
	Status    string   `json:"status"`
}

// RawHunk is the LLM JSON shape before validation.
type RawHunk struct {
	FieldID   string   `json:"field_id"`
	Section   string   `json:"section"`
	Before    string   `json:"before"`
	After     string   `json:"after"`
	Rationale string   `json:"rationale"`
	Evidence  []string `json:"evidence"`
}
