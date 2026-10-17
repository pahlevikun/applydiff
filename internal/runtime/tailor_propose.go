package runtime

import (
	"context"
	"encoding/json"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/store"
	"github.com/pahlevikun/applydiff/internal/tailor"
)

// TailorContext is returned to host agents for external drafting.
type TailorContext struct {
	Job    job.Job       `json:"job"`
	Fields []cv.FieldView `json:"fields"`
	Facts  *cv.Facts     `json:"facts,omitempty"`
	Rules  string        `json:"rules"`
	Schema string        `json:"schema"`
}

// TailorContext builds prompt metadata without calling the LLM.
func (a *App) TailorContext(ctx context.Context, key string) (*TailorContext, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, err
	}
	master, err := a.MasterCV(ctx)
	if err != nil {
		return nil, err
	}
	facts := a.FactsOrFile(ctx)
	prompt := tailor.BuildPrompt(*j, master, facts)
	schema := `{"hunks":[{"field_id":"string","section":"string","before":"string","after":"string","rationale":"string","evidence":["string"]}]}`
	return &TailorContext{
		Job:    *j,
		Fields: master.ListFieldViews(),
		Facts:  facts,
		Rules:  prompt,
		Schema: schema,
	}, nil
}

// ProposeHunks validates host-supplied hunks and saves a proposal.
func (a *App) ProposeHunks(ctx context.Context, key string, raw []tailor.RawHunk, actor Actor) (*store.Proposal, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, err
	}
	master, err := a.MasterCV(ctx)
	if err != nil {
		return nil, err
	}
	facts := a.FactsOrFile(ctx)
	hunks, err := tailor.ValidateHunks(raw, master, facts, j.Description)
	if err != nil {
		return nil, err
	}
	prop, err := a.Store.SaveProposal(ctx, j.ID, hunks)
	if err != nil {
		return nil, err
	}
	if j.Status == "discovered" || j.Status == "matched" {
		_ = a.Store.UpdateJobStatus(ctx, j.ID, "tailoring")
	}
	a.recordAction(ctx, j.ID, "propose", actor, jsonSummary(len(hunks)))
	return prop, nil
}

func jsonSummary(n int) string {
	b, _ := json.Marshal(map[string]int{"hunks": n})
	return string(b)
}
