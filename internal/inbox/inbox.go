package inbox

import (
	"context"
	"strings"
	"time"

	"github.com/pahlevikun/applydiff/internal/store"
)

// Item is one queue row exposed via API/CLI.
type Item struct {
	JobKey    string    `json:"job_key"`
	Title     string    `json:"title"`
	Company   string    `json:"company"`
	State     State     `json:"state"`
	Reason    string    `json:"reason"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Counts tallies items per state.
type Counts struct {
	NeedsDecision int `json:"needs_decision"`
	InProgress    int `json:"in_progress"`
	Sent          int `json:"sent"`
	Interview     int `json:"interview"`
}

// Result is the inbox list plus meta counts.
type Result struct {
	Data  []Item  `json:"data"`
	Meta  Meta    `json:"meta"`
}

type Meta struct {
	Counts Counts `json:"counts"`
}

// ListOptions filters the derived queue.
type ListOptions struct {
	State string
	Limit int
}

// List builds the inbox from jobs + applications + tailoring state (no queue_items table).
func List(ctx context.Context, st *store.Store, opt ListOptions) (Result, error) {
	rows, err := st.ListInboxRows(ctx)
	if err != nil {
		return Result{}, err
	}
	var items []Item
	counts := Counts{}
	for _, row := range rows {
		st, reason := Classify(row)
		if st == "" {
			continue
		}
		switch st {
		case StateNeedsDecision:
			counts.NeedsDecision++
		case StateInProgress:
			counts.InProgress++
		case StateSent:
			counts.Sent++
		case StateInterview:
			counts.Interview++
		}
		if filterState(opt.State, st) {
			items = append(items, Item{
				JobKey: row.Job.ID, Title: row.Job.Title, Company: row.Job.Company,
				State: st, Reason: reason, UpdatedAt: row.Job.UpdatedAt,
			})
		}
	}
	if opt.Limit > 0 && len(items) > opt.Limit {
		items = items[:opt.Limit]
	}
	return Result{Data: items, Meta: Meta{Counts: counts}}, nil
}

func filterState(want string, have State) bool {
	want = strings.TrimSpace(want)
	if want == "" {
		return true
	}
	return strings.EqualFold(want, string(have))
}
