package notify_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/pahlevikun/applydiff/internal/inbox"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/match"
	"github.com/pahlevikun/applydiff/internal/notify"
	"github.com/pahlevikun/applydiff/internal/store"
)

type countNotifier struct {
	n    int
	last notify.Message
}

func (c *countNotifier) Send(_ context.Context, msg notify.Message) error {
	c.n++
	c.last = msg
	return nil
}

func testStore(t *testing.T) *store.Store {
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	return st
}

func seedNeedsDecisionJob(t *testing.T, st *store.Store, id, title string) {
	ctx := context.Background()
	if err := st.InsertJob(ctx, job.Job{
		ID: id, Company: "Co", Title: title, URL: "https://example.com/" + id,
		Source: "test", PostedAt: "2026-10-10T12:00:00Z", Remote: "yes", Region: "global", Status: "discovered",
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.SaveMatchResult(ctx, match.Result{
		JobID: id, Total: inbox.MatchMinScore, Requirements: nil,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestDigestSkippedWhenNothingChanged(t *testing.T) {
	st := testStore(t)
	defer st.Close()
	seedNeedsDecisionJob(t, st, "j1", "Role A")
	ctx := context.Background()
	n := &countNotifier{}
	cfg := notify.Config{BaseURL: "http://localhost:8080"}
	if err := notify.SendNeedsDecisionDigest(ctx, st, n, cfg, "run-1", 1); err != nil {
		t.Fatal(err)
	}
	if n.n != 0 {
		t.Fatalf("expected no send, got %d", n.n)
	}
}

func TestDigestFiresWhenNeedsDecisionGrows(t *testing.T) {
	st := testStore(t)
	defer st.Close()
	seedNeedsDecisionJob(t, st, "j1", "Role A")
	ctx := context.Background()
	n := &countNotifier{}
	cfg := notify.Config{BaseURL: "http://localhost:8080"}
	if err := notify.SendNeedsDecisionDigest(ctx, st, n, cfg, "run-2", 0); err != nil {
		t.Fatal(err)
	}
	if n.n != 1 {
		t.Fatalf("expected one send, got %d", n.n)
	}
	if !strings.Contains(n.last.Title, "1 applications need your decision") {
		t.Fatalf("unexpected title: %q", n.last.Title)
	}
	cnt, err := st.CountNotifications(ctx, store.KindNeedsDecisionDigest)
	if err != nil || cnt != 1 {
		t.Fatalf("notification row: cnt=%d err=%v", cnt, err)
	}
}

func TestDuplicateRunIDDoesNotSendTwice(t *testing.T) {
	st := testStore(t)
	defer st.Close()
	seedNeedsDecisionJob(t, st, "j1", "Role A")
	ctx := context.Background()
	n := &countNotifier{}
	cfg := notify.Config{BaseURL: "http://localhost:8080"}
	runID := "run-dup"
	if err := notify.SendNeedsDecisionDigest(ctx, st, n, cfg, runID, 0); err != nil {
		t.Fatal(err)
	}
	if err := notify.SendNeedsDecisionDigest(ctx, st, n, cfg, runID, 0); err != nil {
		t.Fatal(err)
	}
	if n.n != 1 {
		t.Fatalf("expected one send, got %d", n.n)
	}
}

func TestStdoutNotifierContainsCount(t *testing.T) {
	st := testStore(t)
	defer st.Close()
	seedNeedsDecisionJob(t, st, "j1", "Role A")
	seedNeedsDecisionJob(t, st, "j2", "Role B")
	var buf bytes.Buffer
	ctx := context.Background()
	cfg := notify.Config{BaseURL: "http://localhost:8080"}
	if err := notify.SendNeedsDecisionDigest(ctx, st, notify.Stdout{Out: &buf}, cfg, "run-stdout", 0); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "2 applications need your decision") {
		t.Fatalf("stdout missing count: %q", out)
	}
}

func TestTelegramDisabledWithoutToken(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	st := testStore(t)
	defer st.Close()
	seedNeedsDecisionJob(t, st, "j1", "Role A")
	ctx := context.Background()
	n := notify.DigestNotifier()
	if err := notify.SendNeedsDecisionDigest(ctx, st, n, notify.Config{BaseURL: "http://x"}, "run-nop", 0); err != nil {
		t.Fatal(err)
	}
}
