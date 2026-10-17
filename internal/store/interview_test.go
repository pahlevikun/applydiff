package store

import (
	"context"
	"testing"

	"github.com/pahlevikun/applydiff/internal/interview"
	"github.com/pahlevikun/applydiff/internal/job"
)

func TestInterviewStoreCRUD(t *testing.T) {
	ctx := context.Background()
	st, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	jobID := "job-interview-1"
	if err := st.InsertJob(ctx, job.Job{
		ID: jobID, Company: "Co", Title: "Eng", URL: "https://x", Source: "s",
		PostedAt: "2026-01-01", Status: "discovered",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateApplication(ctx, jobID); err != nil {
		t.Fatal(err)
	}

	kit := &interview.Kit{
		ApplicationID: jobID,
		LoopJSON:      `["a"]`,
		TopicsJSON:    `["b"]`,
		PromptsJSON:   `[]`,
		QuestionsJSON: `[]`,
	}
	if err := st.SaveInterviewKit(ctx, kit); err != nil {
		t.Fatal(err)
	}
	gotKit, err := st.GetInterviewKit(ctx, kit.ID)
	if err != nil || gotKit.ApplicationID != jobID {
		t.Fatalf("kit get: %v %+v", err, gotKit)
	}

	sess := &interview.MockSession{ApplicationID: jobID, Type: "coding", Prompt: "p"}
	steps := []interview.MockStep{{Idx: 0, Name: "Clarify", TimeboxS: 300}}
	if err := st.CreateMockSession(ctx, sess, steps); err != nil {
		t.Fatal(err)
	}
	loadedSteps, err := st.ListMockSteps(ctx, sess.ID)
	if err != nil || len(loadedSteps) != 1 {
		t.Fatalf("steps: %v %d", err, len(loadedSteps))
	}
	loadedSteps[0].Output = "done"
	if err := st.UpdateMockStep(ctx, &loadedSteps[0]); err != nil {
		t.Fatal(err)
	}

	sc := &interview.RubricScore{SessionID: sess.ID, Dimension: "communication", Score: 4}
	if err := st.SaveRubricScore(ctx, sc); err != nil {
		t.Fatal(err)
	}
	scores, err := st.ListRubricScores(ctx, sess.ID)
	if err != nil || len(scores) != 1 {
		t.Fatalf("scores: %v", scores)
	}

	kits, err := st.ListInterviewKits(ctx, jobID)
	if err != nil || len(kits) != 1 {
		t.Fatalf("list kits: %v", kits)
	}
	sessions, err := st.ListMockSessions(ctx, jobID)
	if err != nil || len(sessions) != 1 {
		t.Fatalf("list sessions: %v", sessions)
	}
}

func TestInterviewMigration20260101000007(t *testing.T) {
	ctx := context.Background()
	st, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.ensureMigrationsTable(ctx); err != nil {
		t.Fatal(err)
	}
	raw, err := migrationFS.ReadFile("migrations/20260101000007_interview.sql")
	if err != nil {
		t.Fatal(err)
	}
	for run := 1; run <= 2; run++ {
		if _, err := st.db.ExecContext(ctx, string(raw)); err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
	}
	var n int
	if err := st.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM topic_note WHERE tier = 'mindset'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 7 {
		t.Fatalf("mindset notes %d", n)
	}
	if err := st.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM topic_note WHERE tier = 'reference'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 13 {
		t.Fatalf("reference notes %d", n)
	}
}
