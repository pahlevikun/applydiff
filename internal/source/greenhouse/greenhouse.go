package greenhouse

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
)

const SourceID = "greenhouse"

type Board struct {
	Token   string
	Company string
}

type Source struct {
	Boards []Board
	Client HTTPDoer
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewFromConfig(cfg jobconfig.GreenhouseConfig) *Source {
	boards := make([]Board, 0, len(cfg.Boards))
	for _, b := range cfg.Boards {
		boards = append(boards, Board{Token: b.Token, Company: b.Company})
	}
	return &Source{Boards: boards}
}

func (s *Source) ID() string {
	return SourceID
}

type apiResponse struct {
	Jobs []apiJob `json:"jobs"`
}

type apiJob struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	UpdatedAt   string `json:"updated_at"`
	AbsoluteURL string `json:"absolute_url"`
	Location    struct {
		Name string `json:"name"`
	} `json:"location"`
}

func (s *Source) Fetch(ctx context.Context, queries []string) ([]job.Job, error) {
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	var out []job.Job
	for _, board := range s.Boards {
		if board.Token == "" {
			continue
		}
		jobs, err := s.fetchBoard(ctx, client, board)
		if err != nil {
			return nil, fmt.Errorf("greenhouse board %s: %w", board.Token, err)
		}
		for _, j := range jobs {
			if matchesQueries(j.Title, queries) {
				out = append(out, j)
			}
		}
	}
	return out, nil
}

func (s *Source) fetchBoard(ctx context.Context, client HTTPDoer, board Board) ([]job.Job, error) {
	url := fmt.Sprintf("https://boards-api.greenhouse.io/v1/boards/%s/jobs", board.Token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
	return parseJobs(resp.Body, board)
}

func parseJobs(r io.Reader, board Board) ([]job.Job, error) {
	var payload apiResponse
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		return nil, err
	}
	company := board.Company
	if company == "" {
		company = board.Token
	}
	out := make([]job.Job, 0, len(payload.Jobs))
	for _, j := range payload.Jobs {
		loc := strings.TrimSpace(j.Location.Name)
		out = append(out, job.Job{
			Company:  company,
			Title:    strings.TrimSpace(j.Title),
			URL:      strings.TrimSpace(j.AbsoluteURL),
			Source:   SourceID,
			PostedAt: j.UpdatedAt,
			Remote:   remoteLabel(loc),
			Region:   inferRegion(loc),
			Status:   "discovered",
		})
	}
	return out, nil
}

func matchesQueries(title string, queries []string) bool {
	if len(queries) == 0 {
		return true
	}
	lower := strings.ToLower(title)
	for _, q := range queries {
		q = strings.TrimSpace(strings.ToLower(q))
		if q == "" {
			continue
		}
		if strings.Contains(lower, q) {
			return true
		}
	}
	return false
}

func remoteLabel(location string) string {
	l := strings.ToLower(location)
	if strings.Contains(l, "remote") {
		return "yes"
	}
	return "no"
}

func inferRegion(location string) string {
	l := strings.ToLower(location)
	switch {
	case strings.Contains(l, "singapore"), strings.Contains(l, "jakarta"), strings.Contains(l, "indonesia"),
		strings.Contains(l, "malaysia"), strings.Contains(l, "vietnam"), strings.Contains(l, "thailand"),
		strings.Contains(l, "philippines"), strings.Contains(l, "sea"):
		return "SEA"
	case strings.Contains(l, "europe"), strings.Contains(l, "eu"), strings.Contains(l, "london"),
		strings.Contains(l, "berlin"), strings.Contains(l, "amsterdam"):
		return "EU"
	default:
		return "global"
	}
}
