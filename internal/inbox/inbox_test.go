package inbox_test

import (
	"testing"
	"time"

	"github.com/pahlevikun/applydiff/internal/application"
	"github.com/pahlevikun/applydiff/internal/inbox"
	"github.com/pahlevikun/applydiff/internal/store"
)

func TestClassifyPriorityAndStates(t *testing.T) {
	base := store.InboxRow{
		Job: store.InboxJob{
			ID: "j1", Company: "Acme", Title: "Engineer", Status: "matched",
			UpdatedAt: time.Now().UTC(),
		},
	}

	cases := []struct {
		name   string
		row    store.InboxRow
		want   inbox.State
		reason string
	}{
		{
			name: "needs-decision beats in-progress when shortlisted with pending hunks",
			row: func() store.InboxRow {
				r := base
				r.AppStatus = application.StatusShortlisted
				r.MatchTotal = 20
				r.HasProposal = true
				r.PendingHunks = 3
				return r
			}(),
			want: inbox.StateNeedsDecision,
		},
		{
			name: "in-progress tailoring",
			row: func() store.InboxRow {
				r := base
				r.AppStatus = application.StatusTailoring
				r.HasProposal = true
				r.PendingHunks = 2
				r.MatchTotal = 20
				return r
			}(),
			want: inbox.StateInProgress,
		},
		{
			name: "sent",
			row: func() store.InboxRow {
				r := base
				r.AppStatus = application.StatusSubmitted
				return r
			}(),
			want: inbox.StateSent,
		},
		{
			name: "interview",
			row: func() store.InboxRow {
				r := base
				r.AppStatus = application.StatusFollowup
				return r
			}(),
			want: inbox.StateInterview,
		},
		{
			name: "skipped job excluded",
			row: func() store.InboxRow {
				r := base
				r.Job.Status = "skipped"
				r.MatchTotal = 20
				return r
			}(),
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := inbox.Classify(tc.row)
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestClassifyPriorityOrder(t *testing.T) {
	row := store.InboxRow{Job: store.InboxJob{Status: "matched"}}
	row.AppStatus = application.StatusMatched
	row.MatchTotal = 15
	row.HasProposal = true
	row.PendingHunks = 1
	st, _ := inbox.Classify(row)
	if st != inbox.StateNeedsDecision {
		t.Fatalf("expected needs-decision to win, got %q", st)
	}
}
