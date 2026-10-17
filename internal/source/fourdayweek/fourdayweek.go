package fourdayweek

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

const SourceID = "fourdayweek"

const apiBase = "https://4dayweek.io/api/jobs"

const maxPages = 5
const pageLimit = 100

type Source struct {
	Client HTTPDoer
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewFromConfig(_ jobconfig.FourDayWeekConfig) *Source {
	return &Source{}
}

func (s *Source) ID() string { return SourceID }

type apiResponse struct {
	Jobs    []apiJob `json:"jobs"`
	HasMore bool     `json:"has_more"`
	Page    int      `json:"page"`
}

type apiJob struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	CompanyName     string `json:"company_name"`
	WorkArrangement string `json:"work_arrangement"`
	ScheduleType    string `json:"schedule_type"`
	Salary          string `json:"salary"`
	Category        string `json:"category"`
	Level           string `json:"level"`
	Inserted        string `json:"inserted"`
	IsExpired       bool   `json:"is_expired"`
	Locations       []struct {
		City    string `json:"city"`
		Country string `json:"country"`
	} `json:"locations"`
}

func (s *Source) Fetch(ctx context.Context, queries []string) ([]job.Job, error) {
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Second}
	}
	seen := make(map[string]struct{})
	var out []job.Job
	for page := 1; page <= maxPages; page++ {
		if page > 1 {
			source.PauseBetween(ctx)
		}
		pageJobs, hasMore, err := s.fetchPage(ctx, client, page)
		if err != nil {
			return nil, fmt.Errorf("fourdayweek: %w", err)
		}
		for _, j := range pageJobs {
			if !source.MatchesQueries(queries, j.Title, j.Description, j.Company, strings.Join(j.Keywords, " ")) {
				continue
			}
			k := job.DedupeKey(j)
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			out = append(out, j)
		}
		if !hasMore {
			break
		}
	}
	return out, nil
}

func (s *Source) fetchPage(ctx context.Context, client HTTPDoer, page int) ([]job.Job, bool, error) {
	u := apiBase + "?" + url.Values{
		"page":  {fmt.Sprintf("%d", page)},
		"limit": {fmt.Sprintf("%d", pageLimit)},
	}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, false, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return parsePage(resp.Body)
}

func parsePage(r io.Reader) ([]job.Job, bool, error) {
	var payload apiResponse
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		return nil, false, err
	}
	out := make([]job.Job, 0, len(payload.Jobs))
	for _, j := range payload.Jobs {
		if j.IsExpired {
			continue
		}
		out = append(out, mapJob(j))
	}
	return out, payload.HasMore, nil
}

func mapJob(j apiJob) job.Job {
	company := strings.TrimSpace(j.CompanyName)
	if company == "" {
		company = "4 Day Week"
	}
	loc := formatLocations(j.Locations)
	desc := strings.TrimSpace(strings.Join([]string{
		"Schedule: " + humanSchedule(j.ScheduleType),
		j.Salary,
		loc,
	}, "\n"))
	arr := strings.ToLower(strings.TrimSpace(j.WorkArrangement))
	weekly := scheduleWeeklyDays(j.ScheduleType)
	benefits := []string{"4-day week"}
	if weekly == 4.5 {
		benefits = []string{"4.5-day week"}
	}
	return job.Job{
		Company:         company,
		Title:           strings.TrimSpace(j.Title),
		URL:             "https://4dayweek.io/jobs/" + strings.TrimSpace(j.ID),
		Source:          SourceID,
		PostedAt:        strings.TrimSpace(j.Inserted),
		Remote:          remoteFlag(arr),
		Region:          inferRegion(j.Locations),
		Status:          "discovered",
		Description:     desc,
		SalaryPost:      strings.TrimSpace(j.Salary),
		Arrangement:     arr,
		WeeklyDays:      weekly,
		Benefits:        benefits,
		Keywords:        filterKeywords(j.Category, j.Level),
	}
}

func humanSchedule(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "4_day_week":
		return "4-day week"
	case "4_5_day_week", "4.5_day_week":
		return "4.5-day week"
	default:
		return t
	}
}

func scheduleWeeklyDays(t string) float64 {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "4_5_day_week", "4.5_day_week":
		return 4.5
	default:
		return 4
	}
}

func remoteFlag(arr string) string {
	switch arr {
	case "remote":
		return "yes"
	case "hybrid":
		return "partial"
	default:
		return "no"
	}
}

func formatLocations(locs []struct {
	City    string `json:"city"`
	Country string `json:"country"`
}) string {
	if len(locs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(locs))
	for _, l := range locs {
		city := strings.TrimSpace(l.City)
		country := strings.TrimSpace(l.Country)
		switch {
		case city != "" && country != "":
			parts = append(parts, city+", "+country)
		case country != "":
			parts = append(parts, country)
		case city != "":
			parts = append(parts, city)
		}
	}
	return "Locations: " + strings.Join(parts, "; ")
}

func inferRegion(locs []struct {
	City    string `json:"city"`
	Country string `json:"country"`
}) string {
	for _, l := range locs {
		c := strings.ToLower(l.Country)
		switch {
		case strings.Contains(c, "united kingdom"), strings.Contains(c, "germany"), strings.Contains(c, "france"),
			strings.Contains(c, "netherlands"), strings.Contains(c, "europe"):
			return "EU"
		case strings.Contains(c, "indonesia"), strings.Contains(c, "singapore"), strings.Contains(c, "malaysia"):
			return "SEA"
		case strings.Contains(c, "united states"), strings.Contains(c, "canada"):
			return "global"
		}
	}
	return "global"
}

func filterKeywords(category, level string) []string {
	var out []string
	for _, s := range []string{category, level} {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// PlanURL returns the HTTP URL used for a page (for dry-run plans).
func PlanURL(page int) string {
	if page < 1 {
		page = 1
	}
	return apiBase + "?" + url.Values{
		"page":  {fmt.Sprintf("%d", page)},
		"limit": {fmt.Sprintf("%d", pageLimit)},
	}.Encode()
}
