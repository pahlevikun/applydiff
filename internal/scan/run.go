package scan

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pahlevikun/applydiff/internal/extract"
	"github.com/pahlevikun/applydiff/internal/filter"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/risk"
	"github.com/pahlevikun/applydiff/internal/source"
	"github.com/pahlevikun/applydiff/internal/source/fourdayweek"
	"github.com/pahlevikun/applydiff/internal/store"
)

type FetchPlan struct {
	Source string `json:"source"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

type Result struct {
	Fetched int         `json:"fetched"`
	Deduped int         `json:"deduped"`
	Dead    int         `json:"dead"`
	Stored  int         `json:"stored"`
	Skipped int         `json:"skipped"`
	Plans   []FetchPlan `json:"plans,omitempty"`
}

type Options struct {
	SourceID string
	DryRun   bool
}

type Runner struct {
	Store  *store.Store
	Config *jobconfig.Config
	Client job.HTTPDoer
	MinRisk *int
	MaxRisk *int
	Preset  string
}

func (r *Runner) Run(ctx context.Context, opt Options) (Result, error) {
	var res Result
	if r.Config == nil || r.Store == nil {
		return res, nil
	}

	if opt.DryRun {
		res.Plans = BuildFetchPlans(r.Config, opt.SourceID)
		return res, nil
	}

	spec := r.Config.Filters.EffectiveSpec()
	if strings.EqualFold(strings.TrimSpace(r.Preset), "aggressive") {
		aggressive := filter.AggressivePreset()
		if aggressive.PostedWithin != nil {
			spec.PostedWithin = aggressive.PostedWithin
		}
		if aggressive.MaxApplicants != nil {
			spec.MaxApplicants = aggressive.MaxApplicants
		}
	}
	hashIndex, _ := r.Store.ListJobDescriptionHashes(ctx)

	var fetched []job.Job
	for _, src := range source.List() {
		if opt.SourceID != "" && src.ID() != opt.SourceID {
			continue
		}
		q := queriesFor(r.Config, src.ID())
		jobs, err := src.Fetch(ctx, q)
		srcResult := store.SourceScanResult{
			At:      time.Now().UTC().Format(time.RFC3339),
			Fetched: len(jobs),
		}
		if err != nil {
			srcResult.Error = err.Error()
			_ = r.Store.SaveSourceScanResult(ctx, src.ID(), srcResult)
			return res, err
		}
		if cap := r.Config.Caps.PerSource; cap > 0 && len(jobs) > cap {
			jobs = jobs[:cap]
		}
		srcResult.Fetched = len(jobs)
		fetched = append(fetched, jobs...)
		_ = r.Store.SaveSourceScanResult(ctx, src.ID(), srcResult)
	}
	res.Fetched = len(fetched)

	deduped := job.MergeByDedupeKey(fetched)
	res.Deduped = len(deduped)

	candidates := deduped
	if cap := r.Config.Caps.MaxJobsPerScan; cap > 0 && len(candidates) > cap {
		candidates = candidates[:cap]
	}

	for _, j := range candidates {
		if j.Description != "" || j.Title != "" {
			extract.FromText(&j)
		}
		peerHashes := peerHashesForJob(hashIndex, j.ID)
		assessment := risk.Assess(j, risk.Context{PeerDescriptionHashes: peerHashes})
		j.Risk = assessment.Score
		j.FlagsJSON = assessment.FlagsJSON()
		j.Flags = flagsToJob(assessment.Flags)
		if r.MaxRisk != nil && j.Risk > *r.MaxRisk {
			j.Status = "skipped"
			j.FilterSkipReason = fmt.Sprintf("skipped: risk %d above max %d", j.Risk, *r.MaxRisk)
		}
		if r.MinRisk != nil && j.Risk < *r.MinRisk {
			j.Status = "skipped"
			j.FilterSkipReason = fmt.Sprintf("skipped: risk %d below min %d", j.Risk, *r.MinRisk)
		}
		riskSkipped := j.Status == "skipped"
		eval := filter.Evaluate(j, spec)
		if !eval.Keep && !riskSkipped {
			j.Status = "skipped"
			j.FilterSkipReason = firstSkipReason(eval.Reasons)
		} else if j.Status == "" {
			j.Status = "discovered"
		}

		if os.Getenv("APPLYDIFF_SKIP_JOB_LIVENESS") != "1" && !job.Alive(ctx, r.Client, j.URL) {
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

func peerHashesForJob(index map[string][]string, jobID string) map[string]bool {
	out := map[string]bool{}
	for h, ids := range index {
		for _, id := range ids {
			if id != jobID {
				out[h] = true
				break
			}
		}
	}
	return out
}

func flagsToJob(flags []risk.Flag) []job.JobRiskFlag {
	out := make([]job.JobRiskFlag, 0, len(flags))
	for _, f := range flags {
		out = append(out, job.JobRiskFlag{Code: f.Code, Severity: f.Severity, Message: f.Message})
	}
	return out
}

func queriesFor(cfg *jobconfig.Config, sourceID string) []string {
	if cfg == nil {
		return nil
	}
	switch sourceID {
	case "greenhouse":
		return cfg.QueriesForSource(sourceID, cfg.Sources.Greenhouse.Queries)
	case "remotive":
		return cfg.QueriesForSource(sourceID, cfg.Sources.Remotive.Queries)
	case "remoteok":
		return cfg.QueriesForSource(sourceID, cfg.Sources.RemoteOK.Queries)
	case "jobicy":
		return cfg.QueriesForSource(sourceID, cfg.Sources.Jobicy.Queries)
	case "weworkremotely":
		return cfg.QueriesForSource(sourceID, cfg.Sources.WeWorkRemotely.Queries)
	case "workingnomads":
		return cfg.QueriesForSource(sourceID, cfg.Sources.WorkingNomads.Queries)
	case "fourdayweek":
		return cfg.QueriesForSource(sourceID, cfg.Sources.FourDayWeek.Queries)
	default:
		return cfg.Queries
	}
}

// BuildFetchPlans describes outbound HTTP calls without hitting the network.
func BuildFetchPlans(cfg *jobconfig.Config, onlySource string) []FetchPlan {
	if cfg == nil {
		return nil
	}
	var plans []FetchPlan
	add := func(id string, enabled bool, urls ...string) {
		if !enabled || (onlySource != "" && onlySource != id) {
			return
		}
		for _, u := range urls {
			if u == "" {
				continue
			}
			plans = append(plans, FetchPlan{Source: id, Method: "GET", URL: u})
		}
	}
	if cfg.Sources.Greenhouse.Enabled {
		for _, b := range cfg.Sources.Greenhouse.Boards {
			if b.Token == "" {
				continue
			}
			add("greenhouse", true, "https://boards-api.greenhouse.io/v1/boards/"+b.Token+"/jobs")
		}
	}
	qs := cfg.Sources.Remotive.Queries
	if len(qs) == 0 {
		qs = cfg.Queries
	}
	if len(qs) == 0 {
		qs = []string{""}
	}
	for _, q := range qs {
		u := "https://remotive.com/api/remote-jobs"
		if strings.TrimSpace(q) != "" {
			u += "?search=" + url.QueryEscape(strings.TrimSpace(q))
		}
		add("remotive", cfg.Sources.Remotive.Enabled, u)
	}
	add("remoteok", cfg.Sources.RemoteOK.Enabled, "https://remoteok.com/api")
	jq := cfg.Sources.Jobicy.Queries
	if len(jq) == 0 {
		jq = cfg.Queries
	}
	if len(jq) == 0 {
		jq = []string{""}
	}
	for _, q := range jq {
		u := "https://jobicy.com/api/v2/remote-jobs"
		if strings.TrimSpace(q) != "" {
			u += "?tag=" + url.QueryEscape(strings.TrimSpace(q))
		}
		add("jobicy", cfg.Sources.Jobicy.Enabled, u)
	}
	feed := cfg.Sources.WeWorkRemotely.FeedURL
	if feed == "" {
		feed = "https://weworkremotely.com/categories/remote-programming-jobs.rss"
	}
	add("weworkremotely", cfg.Sources.WeWorkRemotely.Enabled, feed)
	add("workingnomads", cfg.Sources.WorkingNomads.Enabled, "https://www.workingnomads.com/api/exposed_jobs/")
	if cfg.Sources.FourDayWeek.Enabled {
		for p := 1; p <= 5; p++ {
			add("fourdayweek", true, fourdayweek.PlanURL(p))
		}
	}
	return plans
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
