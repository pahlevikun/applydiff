package tailor

import (
	"context"
	"fmt"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/llm"
)

// Engine runs the drafter pass via an LLM runner.
type Engine struct {
	Runner llm.Runner
}

// Generate calls the runner and validates hunks against the master CV and facts gate.
func (e *Engine) Generate(ctx context.Context, j job.Job, master *cv.Document, facts *cv.Facts) ([]Hunk, error) {
	if e == nil || e.Runner == nil {
		return nil, fmt.Errorf("tailor: runner required")
	}
	if master == nil {
		return nil, fmt.Errorf("tailor: master cv required")
	}
	prompt := BuildPrompt(j, master, facts)
	raw, err := e.Runner.Complete(ctx, prompt)
	if err != nil {
		return nil, err
	}
	parsed, err := ParseHunksJSON(raw)
	if err != nil {
		return nil, err
	}
	return ValidateHunks(parsed, master, facts, j.Description)
}

// ValidateHunks applies field checks, facts gate, and ATS keyword tags.
func ValidateHunks(raw []RawHunk, master *cv.Document, facts *cv.Facts, jd string) ([]Hunk, error) {
	jdTerms := JDKeywords(jd)
	var out []Hunk
	for _, r := range raw {
		h := Hunk{
			FieldID:   r.FieldID,
			Section:   r.Section,
			Before:    r.Before,
			After:     r.After,
			Rationale: r.Rationale,
			Evidence:  r.Evidence,
			Status:    StatusPending,
		}
		current, ok := master.Field(r.FieldID)
		if !ok {
			h.Status = StatusBlocked
			out = append(out, h)
			continue
		}
		if r.Before != "" && r.Before != current {
			h.Before = current
		} else if h.Before == "" {
			h.Before = current
		}
		if facts != nil && len(facts.Facts) > 0 {
			if err := facts.GateNumberIn(h.After); err != nil {
				h.Status = StatusBlocked
				out = append(out, h)
				continue
			}
		}
		if AddsJDKeyword(h.Before, h.After, jdTerms) {
			h.Tags = append(h.Tags, TagATSKeyword)
		}
		out = append(out, h)
	}
	return out, nil
}
