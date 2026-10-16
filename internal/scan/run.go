package scan

import (
	"context"
	"strings"
	"time"

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

	filtered := filterJobs(deduped, r.Config.Filters)
	if cap := r.Config.Caps.MaxJobsPerScan; cap > 0 && len(filtered) > cap {
		filtered = filtered[:cap]
	}

	for _, j := range filtered {
		if !job.Alive(ctx, r.Client, j.URL) {
			res.Dead++
			continue
		}
		if err := r.Store.UpsertJob(ctx, j, job.DedupeKey(j)); err != nil {
			return res, err
		}
		res.Stored++
	}
	res.Skipped = len(filtered) - res.Stored - res.Dead
	return res, nil
}

func filterJobs(jobs []job.Job, f jobconfig.Filters) []job.Job {
	out := make([]job.Job, 0, len(jobs))
	cutoff := time.Now().AddDate(0, 0, -f.MaxPostedDays)
	for _, j := range jobs {
		if f.RemoteOnly && j.Remote != "yes" {
			continue
		}
		if !matchesSeniority(j.Title, f.MinSeniority) {
			continue
		}
		if !matchesRegion(j.Region, f.Regions) {
			continue
		}
		if f.MaxPostedDays > 0 {
			if posted, ok := parsePosted(j.PostedAt); ok && posted.Before(cutoff) {
				continue
			}
		}
		out = append(out, j)
	}
	return out
}

func matchesSeniority(title, min string) bool {
	min = strings.TrimSpace(strings.ToLower(min))
	if min == "" {
		return true
	}
	lower := strings.ToLower(title)
	switch min {
	case "senior":
		for _, kw := range []string{"senior", "staff", "principal", "lead", "sr."} {
			if strings.Contains(lower, kw) {
				return true
			}
		}
		return false
	default:
		return strings.Contains(lower, min)
	}
}

func matchesRegion(region string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	region = strings.ToLower(strings.TrimSpace(region))
	for _, a := range allowed {
		if strings.ToLower(strings.TrimSpace(a)) == region {
			return true
		}
	}
	return false
}

func parsePosted(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
