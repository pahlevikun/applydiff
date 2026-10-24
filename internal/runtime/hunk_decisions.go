package runtime

import (
	"context"
	"fmt"

	"github.com/pahlevikun/applydiff/internal/hunkdecision"
	"github.com/pahlevikun/applydiff/internal/store"
	"github.com/pahlevikun/applydiff/internal/tailor"
)

// HunkReview is one hunk with decision metadata for API responses.
type HunkReview struct {
	Index    int          `json:"index"`
	Hunk     tailor.Hunk  `json:"hunk"`
	Decision string       `json:"decision,omitempty"`
	Text     string       `json:"text,omitempty"`
	Stale    bool         `json:"stale"`
	Decided  bool         `json:"decided"`
}

// HunkReviewSummary is returned by GET /api/jobs/{key}/hunks.
type HunkReviewSummary struct {
	JobID   string       `json:"job_id"`
	Total   int          `json:"total"`
	Decided int          `json:"decided"`
	Hunks   []HunkReview `json:"hunks"`
}

// ListHunkReviews loads proposal hunks with persisted decisions.
func (a *App) ListHunkReviews(ctx context.Context, key string) (*HunkReviewSummary, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, err
	}
	prop, err := a.LoadProposal(ctx, key)
	if err != nil {
		return nil, err
	}
	records, err := a.Store.ListHunkDecisions(ctx, j.ID)
	if err != nil {
		return nil, err
	}
	byIndex := map[int]hunkdecision.Record{}
	for _, rec := range records {
		byIndex[rec.HunkIndex] = rec
	}
	out := &HunkReviewSummary{JobID: j.ID, Total: len(prop.Hunks)}
	reviews := make([]HunkReview, len(prop.Hunks))
	for i, h := range prop.Hunks {
		rev := HunkReview{Index: i + 1, Hunk: h, Stale: h.Stale, Decision: h.Decision}
		if rec, ok := byIndex[i]; ok {
			rev.Text = rec.Text
		}
		rev.Decided = hunkDecided(h)
		reviews[i] = rev
		if rev.Decided {
			out.Decided++
		}
	}
	out.Hunks = reviews
	return out, nil
}

// SetHunkDecision persists a decision for hunk index n (1-based).
func (a *App) SetHunkDecision(ctx context.Context, key string, index int, decision, text string, dryRun bool, actor Actor) (*HunkReviewSummary, error) {
	if index < 1 {
		return nil, fmt.Errorf("hunk index out of range: %d", index)
	}
	if !hunkdecision.ValidDecision(decision) {
		return nil, fmt.Errorf("invalid decision %q", decision)
	}
	if decision == hunkdecision.DecisionSkip && text == "" {
		return nil, fmt.Errorf("skip requires a reason in text")
	}
	prop, err := a.LoadProposal(ctx, key)
	if err != nil {
		return nil, err
	}
	if index > len(prop.Hunks) {
		return nil, fmt.Errorf("hunk index out of range: %d", index)
	}
	h := prop.Hunks[index-1]
	if h.Status == tailor.StatusBlocked {
		return nil, fmt.Errorf("hunk %d blocked by facts gate", index)
	}
	hash := hunkdecision.ContentHash(h)
	if dryRun {
		return a.ListHunkReviews(ctx, key)
	}
	_, err = a.Store.UpsertHunkDecision(ctx, prop.JobID, index-1, decision, text, hash)
	if err != nil {
		return nil, err
	}
	if err := a.syncHunkStatusFromDecision(ctx, h.ID, h, decision, text); err != nil {
		return nil, err
	}
	a.recordAction(ctx, prop.JobID, "hunk_"+decision, actor, fmt.Sprintf("%d", index))
	return a.ListHunkReviews(ctx, key)
}

func (a *App) syncHunkStatusFromDecision(ctx context.Context, hunkID string, h tailor.Hunk, decision, text string) error {
	rec := hunkdecision.Record{Decision: decision, Text: text}
	if err := hunkdecision.ApplyToHunk(&h, rec, false); err != nil {
		return err
	}
	return a.Store.SetHunkStatus(ctx, hunkID, h.Status)
}

func hunkDecided(h tailor.Hunk) bool {
	if h.Stale {
		return false
	}
	switch h.Status {
	case tailor.StatusAccepted, tailor.StatusRejected, tailor.StatusSkipped:
		return true
	default:
		return false
	}
}

// UndecidedHunkIndexes returns 1-based indexes that still need a decision.
func UndecidedHunkIndexes(prop *store.Proposal) []int {
	if prop == nil {
		return nil
	}
	var out []int
	for i, h := range prop.Hunks {
		if !hunkDecided(h) {
			out = append(out, i+1)
		}
	}
	return out
}
