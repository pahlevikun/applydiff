package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pahlevikun/applydiff/internal/extract"
	"github.com/pahlevikun/applydiff/internal/job"
)

// AddJobInput is pasted job text or URL metadata from agents.
type AddJobInput struct {
	Title       string
	Company     string
	URL         string
	Description string
	Source      string
}

// AddJob stores a manual job, runs extract.FromText and risk assessment.
func (a *App) AddJob(ctx context.Context, in AddJobInput, actor Actor) (*job.Job, error) {
	if strings.TrimSpace(in.Description) == "" && strings.TrimSpace(in.URL) == "" {
		return nil, fmt.Errorf("add job: description or url required")
	}
	j := job.Job{
		ID:          uuid.NewString(),
		Title:       strings.TrimSpace(in.Title),
		Company:     strings.TrimSpace(in.Company),
		URL:         strings.TrimSpace(in.URL),
		Description: strings.TrimSpace(in.Description),
		Source:      in.Source,
		Status:      "discovered",
	}
	if j.Source == "" {
		j.Source = "manual"
	}
	if j.Title == "" {
		j.Title = "Untitled role"
	}
	if j.Company == "" {
		j.Company = "Unknown"
	}
	extract.FromText(&j)
	peers := a.peerHashesExcluding(ctx, "")
	a.assessAndPersistJobRisk(ctx, &j, peers)
	if err := a.Store.UpsertJob(ctx, j, job.DedupeKey(j)); err != nil {
		return nil, err
	}
	stored, err := a.Store.GetJob(ctx, j.ID)
	if err != nil {
		return nil, err
	}
	a.recordAction(ctx, stored.ID, "add_job", actor, stored.Title)
	return stored, nil
}
