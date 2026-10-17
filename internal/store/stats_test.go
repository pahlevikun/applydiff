package store

import (
	"context"
	"testing"

	"github.com/pahlevikun/applydiff/internal/application"
)

func TestChannelStatsMixedChannels(t *testing.T) {
	ctx := context.Background()
	st, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	jobs := []string{"j1", "j2", "j3"}
	for _, id := range jobs {
		if _, err := st.CreateApplication(ctx, id); err != nil {
			t.Fatal(err)
		}
	}
	if err := st.SetApplicationChannel(ctx, "j1", "referral", "alice", ""); err != nil {
		t.Fatal(err)
	}
	if err := st.SetApplicationChannel(ctx, "j2", "board", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := st.SetApplicationChannel(ctx, "j3", "linkedin", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := st.TransitionApplication(ctx, "j1", application.StatusShortlisted, ""); err != nil {
		t.Fatal(err)
	}
	path := []string{
		application.StatusMatched,
		application.StatusShortlisted,
		application.StatusTailoring,
		application.StatusReady,
		application.StatusSubmitted,
		application.StatusFollowup,
	}
	for _, stName := range path {
		if _, err := st.TransitionApplication(ctx, "j2", stName, ""); err != nil {
			t.Fatal(err)
		}
	}

	stats, err := st.ChannelStats(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) < 2 {
		t.Fatalf("expected multiple channels, got %v", stats)
	}
	var board application.ChannelStats
	for _, s := range stats {
		if s.Channel == "board" {
			board = s
		}
	}
	if board.Total != 1 || board.Replies != 1 || board.InterviewRate != 1 {
		t.Fatalf("board stats: %+v", board)
	}
}
