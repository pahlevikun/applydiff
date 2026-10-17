package runs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pahlevikun/applydiff/internal/store"
)

// Run mirrors store.AsyncRun for callers outside store.
type Run struct {
	ID         string    `json:"id"`
	Kind       string    `json:"kind"`
	JobKey     string    `json:"job_key"`
	Status     string    `json:"status"`
	ResultJSON string    `json:"result_json,omitempty"`
	Error      string    `json:"error,omitempty"`
	Created    time.Time `json:"created_at"`
	Updated    time.Time `json:"updated_at"`
}

// Runner persists async work in SQLite.
type Runner struct {
	Store *store.Store
}

func New(st *store.Store) *Runner {
	return &Runner{Store: st}
}

// Recover marks orphaned running rows as failed.
func (r *Runner) Recover(ctx context.Context) error {
	if r == nil || r.Store == nil {
		return nil
	}
	if r.Store.TableMissing(ctx, "async_runs") {
		return nil
	}
	return r.Store.FailStaleRuns(ctx)
}

// Start runs fn in the background and returns a run id immediately.
func (r *Runner) Start(ctx context.Context, kind, jobKey string, fn func(context.Context) (any, error)) (Run, error) {
	if r == nil || r.Store == nil {
		return Run{}, fmt.Errorf("runs: store required")
	}
	now := time.Now().UTC()
	run := Run{
		ID:      uuid.NewString(),
		Kind:    kind,
		JobKey:  jobKey,
		Status:  store.RunStatusRunning,
		Created: now,
		Updated: now,
	}
	if err := r.Store.InsertRun(ctx, toStore(run)); err != nil {
		return Run{}, err
	}
	go func() {
		bg, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		res, err := fn(bg)
		run.Updated = time.Now().UTC()
		if err != nil {
			run.Status = store.RunStatusFailed
			run.Error = err.Error()
		} else {
			run.Status = store.RunStatusDone
			if res != nil {
				raw, _ := json.Marshal(res)
				run.ResultJSON = string(raw)
			}
		}
		_ = r.Store.UpdateRun(context.Background(), toStore(run))
	}()
	return run, nil
}

// Get loads a run by id.
func (r *Runner) Get(ctx context.Context, id string) (Run, error) {
	if r == nil || r.Store == nil {
		return Run{}, fmt.Errorf("runs: store required")
	}
	sr, err := r.Store.GetRun(ctx, id)
	if err != nil {
		return Run{}, err
	}
	return fromStore(sr), nil
}

func toStore(r Run) store.AsyncRun {
	return store.AsyncRun{
		ID: r.ID, Kind: r.Kind, JobKey: r.JobKey, Status: r.Status,
		ResultJSON: r.ResultJSON, Error: r.Error, Created: r.Created, Updated: r.Updated,
	}
}

func fromStore(r store.AsyncRun) Run {
	return Run{
		ID: r.ID, Kind: r.Kind, JobKey: r.JobKey, Status: r.Status,
		ResultJSON: r.ResultJSON, Error: r.Error, Created: r.Created, Updated: r.Updated,
	}
}
