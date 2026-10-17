package jobicy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/source"
)

const SourceID = "jobicy"

const apiBase = "https://jobicy.com/api/v2/remote-jobs"

type Source struct {
	Client HTTPDoer
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewFromConfig(_ jobconfig.JobicyConfig) *Source {
	return &Source{}
}

func (s *Source) ID() string { return SourceID }

type apiResponse struct {
	Jobs []apiJob `json:"jobs"`
}

type apiJob struct {
	URL          string `json:"url"`
	JobTitle     string `json:"jobTitle"`
	CompanyName  string `json:"companyName"`
	JobGeo       string `json:"jobGeo"`
	JobLevel     string `json:"jobLevel"`
	JobType      []string `json:"jobType"`
	JobExcerpt   string `json:"jobExcerpt"`
	JobDescription string `json:"jobDescription"`
	PubDate      string `json:"pubDate"`
}

func (s *Source) Fetch(ctx context.Context, queries []string) ([]job.Job, error) {
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	if len(queries) == 0 {
		queries = []string{""}
	}
	seen := make(map[string]struct{})
	var out []job.Job
	for i, q := range queries {
		if i > 0 {
			source.PauseBetween(ctx)
		}
		jobs, err := s.fetchQuery(ctx, client, strings.TrimSpace(q))
		if err != nil {
			return nil, fmt.Errorf("jobicy: %w", err)
		}
		qList := []string{q}
		if q == "" {
			qList = nil
		}
		for _, j := range jobs {
			if !source.MatchesQueries(qList, j.Title, j.Description, j.Company) {
				continue
			}
			k := job.DedupeKey(j)
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			out = append(out, j)
		}
	}
	return out, nil
}

func (s *Source) fetchQuery(ctx context.Context, client HTTPDoer, query string) ([]job.Job, error) {
	u := apiBase
	if query != "" {
		u = apiBase + "?tag=" + url.QueryEscape(query)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return parseJobs(resp.Body)
}

func parseJobs(r io.Reader) ([]job.Job, error) {
	var payload apiResponse
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		return nil, err
	}
	out := make([]job.Job, 0, len(payload.Jobs))
	for _, j := range payload.Jobs {
		desc := strings.TrimSpace(j.JobDescription)
		if desc == "" {
			desc = strings.TrimSpace(j.JobExcerpt)
		}
		emp := ""
		if len(j.JobType) > 0 {
			emp = strings.ToLower(strings.TrimSpace(j.JobType[0]))
		}
		out = append(out, job.Job{
			Company:     strings.TrimSpace(j.CompanyName),
			Title:       strings.TrimSpace(j.JobTitle),
			URL:         strings.TrimSpace(j.URL),
			Source:      SourceID,
			PostedAt:    strings.TrimSpace(j.PubDate),
			Remote:      "yes",
			Region:      "global",
			Status:      "discovered",
			Description: desc,
			Seniority:   strings.ToLower(strings.TrimSpace(j.JobLevel)),
			Country:     strings.TrimSpace(j.JobGeo),
			EmploymentType: emp,
			Arrangement: "remote",
		})
	}
	return out, nil
}
