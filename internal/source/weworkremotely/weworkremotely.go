package weworkremotely

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/source"
)

const SourceID = "weworkremotely"

const feedURL = "https://weworkremotely.com/categories/remote-programming-jobs.rss"

type Source struct {
	FeedURL string
	Client  HTTPDoer
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewFromConfig(cfg jobconfig.WeWorkRemotelyConfig) *Source {
	url := strings.TrimSpace(cfg.FeedURL)
	if url == "" {
		url = feedURL
	}
	return &Source{FeedURL: url}
}

func (s *Source) ID() string { return SourceID }

type rssFeed struct {
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
	Region      string `xml:"region"`
}

func (s *Source) Fetch(ctx context.Context, queries []string) ([]job.Job, error) {
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.FeedURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("weworkremotely: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weworkremotely: unexpected status %d", resp.StatusCode)
	}
	jobs, err := parseFeed(resp.Body)
	if err != nil {
		return nil, err
	}
	var out []job.Job
	for _, j := range jobs {
		if source.MatchesQueries(queries, j.Title, j.Description, j.Company) {
			out = append(out, j)
		}
	}
	return out, nil
}

func parseFeed(r io.Reader) ([]job.Job, error) {
	var feed rssFeed
	if err := xml.NewDecoder(r).Decode(&feed); err != nil {
		return nil, err
	}
	out := make([]job.Job, 0, len(feed.Channel.Items))
	for _, item := range feed.Channel.Items {
		company, title := splitTitle(item.Title)
		region := strings.TrimSpace(item.Region)
		out = append(out, job.Job{
			Company:     company,
			Title:       title,
			URL:         strings.TrimSpace(item.Link),
			Source:      SourceID,
			PostedAt:    strings.TrimSpace(item.PubDate),
			Remote:      "yes",
			Region:      inferRegion(region),
			Status:      "discovered",
			Description: strings.TrimSpace(item.Description),
			Country:     region,
			Arrangement: "remote",
		})
	}
	return out, nil
}

func splitTitle(raw string) (company, title string) {
	raw = strings.TrimSpace(raw)
	if i := strings.Index(raw, ":"); i >= 0 {
		return strings.TrimSpace(raw[:i]), strings.TrimSpace(raw[i+1:])
	}
	return "", raw
}

func inferRegion(region string) string {
	l := strings.ToLower(region)
	if strings.Contains(l, "anywhere") || strings.Contains(l, "world") {
		return "global"
	}
	return "global"
}
