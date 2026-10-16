package cv

import "fmt"

// Hunk is one accepted tailoring edit on a job branch.
type Hunk struct {
	FieldID   string   `json:"field_id"`
	Before    string   `json:"before"`
	After     string   `json:"after"`
	Rationale string   `json:"rationale"`
	Evidence  []string `json:"evidence"`
}

// ApplyHunks returns a copy of master with hunks applied in order (last wins per field).
func ApplyHunks(master *Document, hunks []Hunk) (*Document, error) {
	if master == nil {
		return nil, fmt.Errorf("cv: master is nil")
	}
	out := master.Clone()
	if out == nil {
		return nil, fmt.Errorf("cv: failed to clone master")
	}
	byField := make(map[string]Hunk)
	for _, h := range hunks {
		byField[h.FieldID] = h
	}
	for id, h := range byField {
		if err := out.SetField(id, h.After); err != nil {
			return nil, err
		}
	}
	return out, nil
}
