package remotive

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

const SourceID = "remotive"

const apiBase = "https://remotive.com/api/remote-jobs"

type Source struct {
	Client HTTPDoer
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewFromConfig(_ jobconfig.RemotiveConfig) *Source {
	return &Source{}
}

func (s *Source) ID() string { return SourceID }

type apiResponse struct {
	Jobs []apiJob `json:"jobs"`
}

type apiJob struct {
	URL                      string   `json:"url"`
	Title                    string   `json:"title"`
	CompanyName              string   `json:"company_name"`
	PublicationDate          string   `json:"publication_date"`
	CandidateRequiredLocation string  `json:"candidate_required_location"`
	Salary                   string   `json:"salary"`
	Description              string   `json:"description"`
	JobType                  string   `json:"job_type"`
	Tags                     []string `json:"tags"`
}

func (s *Source) Fetch(ctx context.Context, queries []string) ([]job.Job, error) {
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	seen := make(map[string]struct{})
	var out []job.Job
	if len(queries) == 0 {
		queries = []string{""}
	}
	for i, q := range queries {
		if i > 0 {
			source.PauseBetween(ctx)
		}
		jobs, err := s.fetchQuery(ctx, client, strings.TrimSpace(q))
		if err != nil {
			return nil, fmt.Errorf("remotive: %w", err)
		}
		qList := []string{q}
		if q == "" {
			qList = nil
		}
		for _, j := range jobs {
			if !source.MatchesQueries(qList, j.Title, j.Description, strings.Join(j.Keywords, " ")) {
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
		u = apiBase + "?search=" + url.QueryEscape(query)
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
		loc := strings.TrimSpace(j.CandidateRequiredLocation)
		out = append(out, job.Job{
			Company:         strings.TrimSpace(j.CompanyName),
			Title:           strings.TrimSpace(j.Title),
			URL:             strings.TrimSpace(j.URL),
			Source:          SourceID,
			PostedAt:        strings.TrimSpace(j.PublicationDate),
			Remote:          "yes",
			Region:          inferRegion(loc),
			Status:          "discovered",
			Description:     strings.TrimSpace(j.Description),
			SalaryPost:      strings.TrimSpace(j.Salary),
			EmploymentType:  normalizeEmployment(j.JobType),
			Keywords:        j.Tags,
			Arrangement:     "remote",
		})
	}
	return out, nil
}

func normalizeEmployment(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	switch t {
	case "full_time":
		return "full-time"
	case "part_time":
		return "part-time"
	case "contract":
		return "contract"
	default:
		return ""
	}
}

func inferRegion(location string) string {
	l := strings.ToLower(location)
	switch {
	case strings.Contains(l, "usa"), strings.Contains(l, "canada"), strings.Contains(l, "world"):
		return "global"
	case strings.Contains(l, "europe"), strings.Contains(l, "eu"), strings.Contains(l, "uk"):
		return "EU"
	case strings.Contains(l, "singapore"), strings.Contains(l, "indonesia"), strings.Contains(l, "sea"):
		return "SEA"
	default:
		return "global"
	}
}

// PlanURL returns the HTTP URL used for a query (empty query = full feed).
func PlanURL(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return apiBase
	}
	return apiBase + "?search=" + url.QueryEscape(query)
}
