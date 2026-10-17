package remoteok

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

const SourceID = "remoteok"

const apiURL = "https://remoteok.com/api"

type Source struct {
	Client HTTPDoer
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewFromConfig(_ jobconfig.RemoteOKConfig) *Source {
	return &Source{}
}

func (s *Source) ID() string { return SourceID }

func (s *Source) Fetch(ctx context.Context, queries []string) ([]job.Job, error) {
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Second}
	}
	all, err := s.fetchAll(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("remoteok: %w", err)
	}
	if len(queries) == 0 {
		return all, nil
	}
	var out []job.Job
	for _, j := range all {
		if source.MatchesQueries(queries, j.Title, j.Description, j.Company) {
			out = append(out, j)
		}
	}
	return out, nil
}

func (s *Source) fetchAll(ctx context.Context, client HTTPDoer) ([]job.Job, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
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
	var raw []map[string]interface{}
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, err
	}
	out := make([]job.Job, 0, len(raw))
	for _, row := range raw {
		if row["id"] == nil && row["position"] == nil {
			continue
		}
		title := stringField(row, "position")
		if title == "" {
			continue
		}
		company := stringField(row, "company")
		url := stringField(row, "url")
		if url == "" {
			slug := stringField(row, "slug")
			if slug != "" {
				url = "https://remoteok.com/remote-jobs/" + slug
			}
		}
		date := stringField(row, "date")
		desc := stringField(row, "description")
		tags := stringSliceField(row, "tags")
		out = append(out, job.Job{
			Company:     company,
			Title:       title,
			URL:         url,
			Source:      SourceID,
			PostedAt:    date,
			Remote:      "yes",
			Region:      "global",
			Status:      "discovered",
			Description: desc,
			Keywords:    tags,
			Arrangement: "remote",
		})
	}
	return out, nil
}

func stringField(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func stringSliceField(m map[string]interface{}, key string) []string {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	switch t := v.(type) {
	case []interface{}:
		var out []string
		for _, item := range t {
			s := strings.TrimSpace(fmt.Sprint(item))
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return t
	default:
		return nil
	}
}
