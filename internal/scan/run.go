package scan

import (
	"context"
	"strings"

	"github.com/pahlevikun/applydiff/internal/extract"
	"github.com/pahlevikun/applydiff/internal/filter"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/source"
	"github.com/pahlevikun/applydiff/internal/store"
)

type Result struct {
	Fetched int `json:"fetched"`
	Deduped int `json:"deduped"`
	Dead    int `json:"dead"`
	Stored  int `json:"stored"`
	Skipped int `json:"skipped"`
}

type Runner struct {
	Store  *store.Store
	Config *jobconfig.Config
	Client job.HTTPDoer
}

func (r *Runner) Run(ctx context.Context) (Result, error) {
	var res Result
	if r.Config == nil || r.Store == nil {
		return res, nil
	}

	spec := r.Config.Filters.EffectiveSpec()

	queries := r.Config.Queries
	var fetched []job.Job
	for _, src := range source.List() {
		jobs, err := src.Fetch(ctx, queries)
		if err != nil {
			return res, err
		}
		if cap := r.Config.Caps.PerSource; cap > 0 && len(jobs) > cap {
			jobs = jobs[:cap]
		}
		fetched = append(fetched, jobs...)
	}
	res.Fetched = len(fetched)

	deduped := job.Dedupe(fetched)
	res.Deduped = len(deduped)

	candidates := deduped
	if cap := r.Config.Caps.MaxJobsPerScan; cap > 0 && len(candidates) > cap {
		candidates = candidates[:cap]
	}

	for _, j := range candidates {
		if j.Description != "" || j.Title != "" {
			extract.FromText(&j)
		}
		eval := filter.Evaluate(j, spec)
		if !eval.Keep {
			j.Status = "skipped"
			j.FilterSkipReason = firstSkipReason(eval.Reasons)
		} else if j.Status == "" {
			j.Status = "discovered"
		}

		if !job.Alive(ctx, r.Client, j.URL) {
			res.Dead++
			continue
		}
		if err := r.Store.UpsertJob(ctx, j, job.DedupeKey(j)); err != nil {
			return res, err
		}
		if j.Status == "skipped" {
			res.Skipped++
		} else {
			res.Stored++
		}
	}
	return res, nil
}

func firstSkipReason(reasons []string) string {
	for _, r := range reasons {
		if strings.HasPrefix(r, "skipped:") {
			return r
		}
	}
	if len(reasons) > 0 {
		return reasons[0]
	}
	return "skipped: filter"
}
