package daemon_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/daemon"
	"github.com/pahlevikun/applydiff/internal/filter"
	"github.com/pahlevikun/applydiff/internal/job"
	"github.com/pahlevikun/applydiff/internal/jobconfig"
	"github.com/pahlevikun/applydiff/internal/llm"
	"github.com/pahlevikun/applydiff/internal/runtime"
	"github.com/pahlevikun/applydiff/internal/scan"
	"github.com/pahlevikun/applydiff/internal/source"
	"github.com/pahlevikun/applydiff/internal/store"
)

type stubSource struct {
	id   string
	jobs []job.Job
	err  error
}

func (s stubSource) ID() string { return s.id }

func (s stubSource) Fetch(ctx context.Context, queries []string) ([]job.Job, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.jobs, nil
}

type aliveClient struct{}

func (aliveClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
}

func testApp(t *testing.T, st *store.Store, sources ...stubSource) *runtime.App {
	source.Reset()
	for _, s := range sources {
		source.Register(s)
	}
	cfg := &jobconfig.Config{
		Filters: jobconfig.Filters{Spec: filter.Spec{
			Regions:       &filter.CategoryRule{Hard: false, Values: []string{"global"}},
			MaxPostedDays: &filter.IntRule{Hard: false, Max: 365},
		}},
	}
	doc := sampleCV()
	if err := st.SaveMaster(context.Background(), doc); err != nil {
		t.Fatal(err)
	}
	return &runtime.App{
		Store:  st,
		Config: cfg,
		Runner: &scan.Runner{Store: st, Config: cfg, Client: aliveClient{}},
		LLM:    llm.FakeRunner{Response: `{"hunks":[{"field_id":"summary","section":"summary","before":"a","after":"b","rationale":"t","evidence":[],"tags":[]}]}`},
	}
}

func sampleCV() *cv.Document {
	end := "2022-12"
	return &cv.Document{
		Person: "test",
		Contact: cv.Contact{Email: "a@b.com", Name: "Test"},
		Summary: cv.Summary{ID: "summary", Text: "Senior Go engineer with distributed systems experience."},
		Experience: []cv.Experience{
			{
				ID: "exp1", Company: "Co", Title: "Engineer", Start: "2020-01", End: &end,
				Bullets: []cv.Bullet{{ID: "b1", Text: "Built APIs in Go"}},
			},
		},
		Skills: []cv.SkillGroup{{ID: "skills", Label: "Skills", Items: []string{"Go", "SQL", "Kubernetes"}}},
	}
}

func TestDaemonNotifyWithoutTelegramStillSucceeds(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	app := testApp(t, st, stubSource{
		id: "good",
		jobs: []job.Job{{
			Company: "Acme", Title: "Senior Go Engineer", URL: "https://example.com/notify-1",
			Source: "good", PostedAt: "2026-10-10", Remote: "yes", Region: "global", Status: "discovered",
			Description: "We need Go, Kubernetes, and backend APIs.",
		}},
	})
	dir := t.TempDir()
	if err := daemon.Run(context.Background(), app, daemon.Options{
		Once: true, DataDir: dir, MaxDraftsPerDay: 10, Notify: true,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestDaemonOnceWritesRun(t *testing.T) {
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	app := testApp(t, st, stubSource{
		id: "good",
		jobs: []job.Job{{
			Company: "Acme", Title: "Senior Go Engineer", URL: "https://example.com/1",
			Source: "good", PostedAt: "2026-10-10", Remote: "yes", Region: "global", Status: "discovered",
			Description: "We need Go, Kubernetes, and backend APIs.",
		}},
	})
	dir := t.TempDir()
	if err := daemon.Run(context.Background(), app, daemon.Options{
		Once: true, DataDir: dir, MaxDraftsPerDay: 10,
	}); err != nil {
		t.Fatal(err)
	}
	n, err := st.CountPipelineRuns(context.Background(), "daemon_tick")
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 run row, got %d", n)
	}
}

func TestPartialSourceFailureStillSucceeds(t *testing.T) {
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	app := testApp(t, st,
		stubSource{id: "bad", err: os.ErrPermission},
		stubSource{id: "good", jobs: []job.Job{{
			Company: "Beta", Title: "Engineer", URL: "https://example.com/2",
			Source: "good", PostedAt: "2026-10-10", Remote: "yes", Region: "global", Status: "discovered",
			Description: "Go backend role.",
		}}},
	)
	dir := t.TempDir()
	if err := daemon.Run(context.Background(), app, daemon.Options{Once: true, DataDir: dir}); err != nil {
		t.Fatal(err)
	}
	n, _ := st.CountPipelineRuns(context.Background(), "daemon_tick")
	if n != 1 {
		t.Fatalf("expected tick to record run, got %d", n)
	}
}

func TestDraftCap(t *testing.T) {
	t.Setenv("LLM_API_KEY", "test-key")
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	var jobs []job.Job
	for i := 0; i < 5; i++ {
		jobs = append(jobs, job.Job{
			Company: "Co", Title: "Go Engineer " + string(rune('A'+i)), URL: "https://example.com/job/" + string(rune('a'+i)),
			Source: "bulk", PostedAt: "2026-10-10", Remote: "yes", Region: "global", Status: "discovered",
			Description: "Senior Go engineer backend APIs Kubernetes.",
		})
	}
	app := testApp(t, st, stubSource{id: "bulk", jobs: jobs})
	dir := t.TempDir()
	if err := daemon.Run(context.Background(), app, daemon.Options{
		Once: true, DataDir: dir, MaxDraftsPerDay: 2,
	}); err != nil {
		t.Fatal(err)
	}
	count, err := st.CountProposalsSince(context.Background(), time.Now().UTC().Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected 2 drafts, got %d", count)
	}
}

func TestSecondDaemonExitsWhenLocked(t *testing.T) {
	dir := t.TempDir()
	release, held, _ := daemon.TryLock(dir)
	if !held {
		t.Fatal("expected lock held")
	}
	defer release()
	_, held2, msg := daemon.TryLock(dir)
	if held2 {
		t.Fatal("expected second lock to fail")
	}
	if msg == "" {
		t.Fatal("expected message")
	}
}

func TestLockStalePID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "daemon.lock")
	_ = os.WriteFile(path, []byte("99999999\n"), 0o644)
	_, held, _ := daemon.TryLock(dir)
	if !held {
		t.Fatal("expected to take stale lock")
	}
}

func TestDaemonLockConcurrent(t *testing.T) {
	dir := t.TempDir()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		release, held, _ := daemon.TryLock(dir)
		if !held {
			t.Error("first lock failed")
			return
		}
		time.Sleep(200 * time.Millisecond)
		release()
	}()
	time.Sleep(50 * time.Millisecond)
	_, held, msg := daemon.TryLock(dir)
	if held {
		t.Fatal("expected block message, got lock")
	}
	if msg == "" {
		t.Fatal("missing message")
	}
	wg.Wait()
}
