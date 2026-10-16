package diff

import (
	"fmt"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/tailor"
)

// Apply applies one hunk onto a CV copy (field becomes after).
func Apply(doc *cv.Document, h tailor.Hunk) (*cv.Document, error) {
	if doc == nil {
		return nil, fmt.Errorf("diff: document is nil")
	}
	out := doc.Clone()
	if out == nil {
		return nil, fmt.Errorf("diff: clone failed")
	}
	if err := out.SetField(h.FieldID, h.After); err != nil {
		return nil, err
	}
	return out, nil
}

// Revert restores the before value for one hunk.
func Revert(doc *cv.Document, h tailor.Hunk) (*cv.Document, error) {
	if doc == nil {
		return nil, fmt.Errorf("diff: document is nil")
	}
	out := doc.Clone()
	if out == nil {
		return nil, fmt.Errorf("diff: clone failed")
	}
	if err := out.SetField(h.FieldID, h.Before); err != nil {
		return nil, err
	}
	return out, nil
}

// ApplyAllAccepted applies accepted hunks onto master (last wins per field).
func ApplyAllAccepted(master *cv.Document, hunks []tailor.Hunk) (*cv.Document, error) {
	if master == nil {
		return nil, fmt.Errorf("diff: master is nil")
	}
	out := master.Clone()
	if out == nil {
		return nil, fmt.Errorf("diff: clone failed")
	}
	byField := make(map[string]tailor.Hunk)
	for _, h := range hunks {
		if h.Status != tailor.StatusAccepted {
			continue
		}
		byField[h.FieldID] = h
	}
	for id, h := range byField {
		if err := out.SetField(id, h.After); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// ToCVHunks converts accepted tailor hunks for legacy cv.ApplyHunks.
func ToCVHunks(hunks []tailor.Hunk) []cv.Hunk {
	var out []cv.Hunk
	for _, h := range hunks {
		if h.Status != tailor.StatusAccepted {
			continue
		}
		out = append(out, cv.Hunk{
			FieldID:   h.FieldID,
			Before:    h.Before,
			After:     h.After,
			Rationale: h.Rationale,
			Evidence:  h.Evidence,
		})
	}
	return out
}
