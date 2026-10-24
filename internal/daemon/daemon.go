package daemon

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pahlevikun/applydiff/internal/inbox"
	"github.com/pahlevikun/applydiff/internal/llm"
	"github.com/pahlevikun/applydiff/internal/notify"
	"github.com/pahlevikun/applydiff/internal/runtime"
	"github.com/pahlevikun/applydiff/internal/scan"
	"github.com/pahlevikun/applydiff/internal/source"
	"github.com/pahlevikun/applydiff/internal/store"

	"github.com/google/uuid"
)

const runKindDaemonTick = "daemon_tick"

// Options configure the background engine loop.
type Options struct {
	Interval        time.Duration
	MaxDraftsPerDay int
	Once            bool
	DryRun          bool
	Notify          bool
	DataDir         string
}

// Run acquires the lock and runs the tick loop until stopped or --once.
func Run(ctx context.Context, app *runtime.App, opt Options) error {
	if app == nil {
		return fmt.Errorf("daemon: app required")
	}
	dataDir := opt.DataDir
	if dataDir == "" {
		dataDir = app.Paths.Home
	}
	release, held, msg := TryLock(dataDir)
	if !held {
		log.Println(msg)
		return nil
	}
	defer release()

	logger, closeLog, err := openRunLog(dataDir)
	if err != nil {
		return err
	}
	defer closeLog()
	log.SetOutput(io.MultiWriter(os.Stderr, logger))

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	runTick := func() {
		if err := tick(ctx, app, opt); err != nil {
			log.Printf("daemon tick error: %v", err)
		}
	}
	runTick()
	if opt.Once {
		return nil
	}
	ticker := time.NewTicker(opt.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("daemon shutdown")
			return nil
		case <-ticker.C:
			runTick()
		}
	}
}

func openRunLog(dataDir string) (*os.File, func(), error) {
	path := filepath.Join(dataDir, "runs.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, func() {}, err
	}
	return f, func() { _ = f.Close() }, nil
}

func tick(ctx context.Context, app *runtime.App, opt Options) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	started := time.Now().UTC()
	needsBefore := 0
	if opt.Notify && !opt.DryRun {
		needsBefore, _ = notify.NeedsDecisionCount(ctx, app.Store)
	}
	detail := store.PipelineRunDetail{DryRun: opt.DryRun, Notify: opt.Notify, Sources: map[string]store.SourceTickResult{}}
	var fetched, stored, skipped int
	var tickErr string

	for _, src := range source.List() {
		if ctx.Err() != nil {
			break
		}
		srcRes := store.SourceTickResult{}
		res, err := app.Runner.Run(ctx, scan.Options{SourceID: src.ID(), DryRun: opt.DryRun})
		srcRes.Fetched = res.Fetched
		srcRes.Stored = res.Stored
		srcRes.Skipped = res.Skipped
		if err != nil {
			srcRes.Error = err.Error()
			log.Printf("source %s failed: %v", src.ID(), err)
		}
		detail.Sources[src.ID()] = srcRes
		fetched += res.Fetched
		stored += res.Stored
		skipped += res.Skipped
	}

	if ctx.Err() == nil && !opt.DryRun {
		matched, err := matchJobs(ctx, app)
		detail.Matched = matched
		if err != nil {
			tickErr = err.Error()
		} else {
			drafted, note := draftJobs(ctx, app, opt.MaxDraftsPerDay)
			detail.Drafted = drafted
			detail.DraftNote = note
		}
	} else if opt.DryRun {
		detail.DraftNote = "dry-run"
	} else if ctx.Err() != nil {
		tickErr = ctx.Err().Error()
	}

	finished := time.Now().UTC()
	if opt.DryRun {
		return nil
	}
	runID := uuid.NewString()
	if err := app.Store.InsertPipelineRun(ctx, store.PipelineRun{
		ID:         runID,
		Kind:       runKindDaemonTick,
		StartedAt:  started,
		FinishedAt: finished,
		Fetched:    fetched,
		Stored:     stored,
		Skipped:    skipped,
		Error:      tickErr,
		DetailJSON: store.MarshalPipelineDetail(detail),
	}); err != nil {
		return err
	}
	if opt.Notify {
		cfg := notify.LoadConfig()
		if err := notify.SendNeedsDecisionDigest(ctx, app.Store, notify.DigestNotifier(), cfg, runID, needsBefore); err != nil {
			log.Printf("notification error: %v", err)
		}
	}
	return nil
}

func matchJobs(ctx context.Context, app *runtime.App) (int, error) {
	ids, err := app.Store.ListJobsNeedingMatch(ctx)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, id := range ids {
		if ctx.Err() != nil {
			return n, ctx.Err()
		}
		res, err := app.ComputeMatch(ctx, id, false)
		if err != nil {
			continue
		}
		if res != nil && res.Total >= inbox.MatchMinScore {
			_ = app.RecordMatchPipeline(ctx, id)
			n++
		}
	}
	return n, nil
}

func draftJobs(ctx context.Context, app *runtime.App, maxPerDay int) (int, string) {
	if !llm.BackendConfigured() {
		return 0, "no_llm"
	}
	since := time.Now().UTC().Add(-24 * time.Hour)
	existing, err := app.Store.CountProposalsSince(ctx, since)
	if err != nil {
		return 0, "count_error"
	}
	if existing >= maxPerDay {
		log.Printf("draft cap reached (%d in last 24h, max %d)", existing, maxPerDay)
		return 0, "cap"
	}
	remaining := maxPerDay - existing
	drafted := 0
	rows, err := app.Store.ListInboxRows(ctx)
	if err != nil {
		return 0, "list_error"
	}
	for _, row := range rows {
		if ctx.Err() != nil {
			break
		}
		if drafted >= remaining {
			break
		}
		if row.HasProposal || row.MatchTotal < inbox.MatchMinScore {
			continue
		}
		if row.AppStatus != "" && row.AppStatus != "discovered" && row.AppStatus != "matched" && row.AppStatus != "shortlisted" {
			continue
		}
		if row.Job.Status == "skipped" || row.Job.FilterSkipReason != "" {
			continue
		}
		_, err := app.Tailor(ctx, row.Job.ID, false)
		if err != nil {
			log.Printf("draft %s: %v", row.Job.ID, err)
			continue
		}
		_ = app.RecordTailorStarted(ctx, row.Job.ID)
		drafted++
	}
	return drafted, ""
}
