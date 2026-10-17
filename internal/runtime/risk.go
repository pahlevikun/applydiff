package runtime

import (
	"context"

	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/risk"
)

func (a *App) assessAndPersistJobRisk(ctx context.Context, j *job.Job, peerHashes map[string]bool) {
	if a == nil || a.Store == nil || j == nil {
		return
	}
	ctxAssess := risk.Context{PeerDescriptionHashes: peerHashes}
	assessment := risk.Assess(*j, ctxAssess)
	j.Risk = assessment.Score
	j.FlagsJSON = assessment.FlagsJSON()
	j.Flags = make([]job.JobRiskFlag, 0, len(assessment.Flags))
	for _, f := range assessment.Flags {
		j.Flags = append(j.Flags, job.JobRiskFlag{Code: f.Code, Severity: f.Severity, Message: f.Message})
	}
	_ = a.Store.UpdateJobRisk(ctx, j.ID, j.Risk, j.FlagsJSON)
}

func (a *App) peerHashesExcluding(ctx context.Context, skipID string) map[string]bool {
	index, err := a.Store.ListJobDescriptionHashes(ctx)
	if err != nil {
		return nil
	}
	out := map[string]bool{}
	for h, ids := range index {
		for _, id := range ids {
			if id != skipID {
				out[h] = true
				break
			}
		}
	}
	return out
}
