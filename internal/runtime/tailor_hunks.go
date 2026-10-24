package runtime

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pahlevikun/applydiff/internal/hunkdecision"
	"github.com/pahlevikun/applydiff/internal/store"
	"github.com/pahlevikun/applydiff/internal/tailor"
)

// Tailor generates a proposal via LLM.
func (a *App) Tailor(ctx context.Context, key string, dryRun bool) (*store.Proposal, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, err
	}
	master, err := a.MasterCV(ctx)
	if err != nil {
		return nil, err
	}
	facts := a.FactsOrFile(ctx)
	eng := &tailor.Engine{Runner: a.LLM}
	hunks, err := eng.Generate(ctx, *j, master, facts)
	if err != nil {
		return nil, ErrLLM{Err: err}
	}
	if dryRun {
		return &store.Proposal{JobID: j.ID, Hunks: hunks}, nil
	}
	prop, err := a.Store.SaveProposal(ctx, j.ID, hunks)
	if err != nil {
		return nil, err
	}
	if j.Status == "discovered" || j.Status == "matched" {
		_ = a.Store.UpdateJobStatus(ctx, j.ID, "tailoring")
	}
	return prop, nil
}

// LoadProposal loads tailoring proposal for a job.
func (a *App) LoadProposal(ctx context.Context, key string) (*store.Proposal, error) {
	j, err := a.JobKey(ctx, key)
	if err != nil {
		return nil, err
	}
	prop, err := a.Store.LoadProposal(ctx, j.ID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no proposal for job %s", key)
	}
	return prop, err
}

// PendingHunks counts hunks still pending, blocked, stale, or skip without reason.
func PendingHunks(hunks []tailor.Hunk) int {
	n := 0
	for _, h := range hunks {
		if h.Stale || h.Status == tailor.StatusPending || h.Status == tailor.StatusBlocked {
			n++
		}
	}
	return n
}

// ApproveHunks sets accepted status on listed indices (1-based) or all with all=true.
func (a *App) ApproveHunks(ctx context.Context, key string, indices []int, all bool, dryRun bool, actor Actor) (*store.Proposal, error) {
	prop, err := a.LoadProposal(ctx, key)
	if err != nil {
		return nil, err
	}
	targets := map[int]bool{}
	if all {
		for i := range prop.Hunks {
			targets[i] = true
		}
	} else {
		for _, n := range indices {
			if n < 1 || n > len(prop.Hunks) {
				return nil, fmt.Errorf("hunk index out of range: %d", n)
			}
			targets[n-1] = true
		}
	}
	if dryRun {
		out := *prop
		for i := range out.Hunks {
			if targets[i] && out.Hunks[i].Status != tailor.StatusBlocked {
				out.Hunks[i].Status = tailor.StatusAccepted
			}
		}
		return &out, nil
	}
	for i, h := range prop.Hunks {
		if !targets[i] {
			continue
		}
		if h.Status == tailor.StatusBlocked {
			return nil, fmt.Errorf("hunk %d blocked by facts gate", i+1)
		}
		if err := a.Store.SetHunkStatus(ctx, h.ID, tailor.StatusAccepted); err != nil {
			return nil, err
		}
		hash := hunkdecision.ContentHash(h)
		_, _ = a.Store.UpsertHunkDecision(ctx, prop.JobID, i, hunkdecision.DecisionAccept, "", hash)
	}
	prop, err = a.LoadProposal(ctx, key)
	if err == nil {
		a.recordAction(ctx, prop.JobID, "approve", actor, "")
	}
	return prop, err
}

// RejectHunks rejects listed indices (1-based).
func (a *App) RejectHunks(ctx context.Context, key string, indices []int, dryRun bool, actor Actor) (*store.Proposal, error) {
	prop, err := a.LoadProposal(ctx, key)
	if err != nil {
		return nil, err
	}
	if dryRun {
		out := *prop
		for _, n := range indices {
			if n < 1 || n > len(out.Hunks) {
				return nil, fmt.Errorf("hunk index out of range: %d", n)
			}
			if out.Hunks[n-1].Status != tailor.StatusBlocked {
				out.Hunks[n-1].Status = tailor.StatusRejected
			}
		}
		return &out, nil
	}
	for _, n := range indices {
		if n < 1 || n > len(prop.Hunks) {
			return nil, fmt.Errorf("hunk index out of range: %d", n)
		}
		h := prop.Hunks[n-1]
		if h.Status == tailor.StatusBlocked {
			continue
		}
		if err := a.Store.SetHunkStatus(ctx, h.ID, tailor.StatusRejected); err != nil {
			return nil, err
		}
		hash := hunkdecision.ContentHash(h)
		_, _ = a.Store.UpsertHunkDecision(ctx, prop.JobID, n-1, hunkdecision.DecisionReject, "", hash)
	}
	prop, err = a.LoadProposal(ctx, key)
	if err == nil {
		a.recordAction(ctx, prop.JobID, "reject", actor, "")
	}
	return prop, err
}

// RequireAllDecided returns ErrNeedsApproval when pending hunks remain.
func RequireAllDecided(prop *store.Proposal) error {
	if prop == nil {
		return nil
	}
	undecided := UndecidedHunkIndexes(prop)
	if len(undecided) > 0 {
		return ErrNeedsApproval{Pending: len(undecided), Undecided: undecided}
	}
	return nil
}

// IsLLMError heuristically classifies errors for exit code 4.
func IsLLMError(err error) bool {
	if err == nil {
		return false
	}
	var le ErrLLM
	if AsErrLLM(err, &le) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "llm") || strings.Contains(msg, "tailor:")
}

func AsErrLLM(err error, target *ErrLLM) bool {
	if e, ok := err.(ErrLLM); ok {
		*target = e
		return true
	}
	return false
}
