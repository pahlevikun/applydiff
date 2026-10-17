package workingnomads

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/source"
)

const SourceID = "workingnomads"

const apiURL = "https://www.workingnomads.com/api/exposed_jobs/"

type Source struct {
	Client HTTPDoer
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewFromConfig(_ jobconfig.WorkingNomadsConfig) *Source {
	return &Source{}
}

func (s *Source) ID() string { return SourceID }

type apiJob struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CompanyName string `json:"company_name"`
	Category    string `json:"category_name"`
	Tags        string `json:"tags"`
	Location    string `json:"location"`
	PubDate     string `json:"pub_date"`
}

func (s *Source) Fetch(ctx context.Context, queries []string) ([]job.Job, error) {
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("workingnomads: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("workingnomads: unexpected status %d", resp.StatusCode)
	}
	jobs, err := parseJobs(resp.Body)
	if err != nil {
		return nil, err
	}
	var out []job.Job
	for _, j := range jobs {
		if source.MatchesQueries(queries, j.Title, j.Description, j.Company, strings.Join(j.Keywords, " ")) {
			out = append(out, j)
		}
	}
	return out, nil
}

func parseJobs(r io.Reader) ([]job.Job, error) {
	var payload []apiJob
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		return nil, err
	}
	out := make([]job.Job, 0, len(payload))
	for _, j := range payload {
		var tags []string
		for _, part := range strings.Split(j.Tags, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				tags = append(tags, part)
			}
		}
		out = append(out, job.Job{
			Company:     strings.TrimSpace(j.CompanyName),
			Title:       strings.TrimSpace(j.Title),
			URL:         strings.TrimSpace(j.URL),
			Source:      SourceID,
			PostedAt:    strings.TrimSpace(j.PubDate),
			Remote:      "yes",
			Region:      "global",
			Status:      "discovered",
			Description: strings.TrimSpace(j.Description),
			Country:     strings.TrimSpace(j.Location),
			Keywords:    tags,
			Archetype:   strings.ToLower(strings.TrimSpace(j.Category)),
			Arrangement: "remote",
		})
	}
	return out, nil
}
