package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pahlevikun/applydiff/internal/interview"
)

func (s *Store) SaveInterviewKit(ctx context.Context, kit *interview.Kit) error {
	if kit == nil || kit.ApplicationID == "" {
		return fmt.Errorf("store: interview kit requires application_id")
	}
	if kit.ID == "" {
		kit.ID = uuid.NewString()
	}
	if kit.CreatedAt == "" {
		kit.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO interview_kit (id, application_id, loop_json, topics_json, prompts_json, questions_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, kit.ID, kit.ApplicationID, kit.LoopJSON, kit.TopicsJSON, kit.PromptsJSON, kit.QuestionsJSON, kit.CreatedAt)
	return err
}

func (s *Store) GetInterviewKit(ctx context.Context, id string) (*interview.Kit, error) {
	var k interview.Kit
	err := s.db.QueryRowContext(ctx, `
		SELECT id, application_id, loop_json, topics_json, prompts_json, questions_json, created_at
		FROM interview_kit WHERE id = ?
	`, id).Scan(&k.ID, &k.ApplicationID, &k.LoopJSON, &k.TopicsJSON, &k.PromptsJSON, &k.QuestionsJSON, &k.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (s *Store) ListInterviewKits(ctx context.Context, applicationID string) ([]interview.Kit, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, application_id, loop_json, topics_json, prompts_json, questions_json, created_at
		FROM interview_kit WHERE application_id = ? ORDER BY created_at DESC
	`, applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []interview.Kit
	for rows.Next() {
		var k interview.Kit
		if err := rows.Scan(&k.ID, &k.ApplicationID, &k.LoopJSON, &k.TopicsJSON, &k.PromptsJSON, &k.QuestionsJSON, &k.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func (s *Store) DeleteInterviewKit(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM interview_kit WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) UpdateStory(ctx context.Context, story *interview.Story) error {
	if story == nil || story.ID == "" {
		return fmt.Errorf("store: story id required")
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE story_bank SET category = ?, situation = ?, task = ?, action = ?, result = ?, learning = ?, evidence_json = ?
		WHERE id = ?
	`, story.Category, story.Situation, story.Task, story.Action, story.Result, story.Learning, story.EvidenceJSON, story.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) FinishMockSession(ctx context.Context, sessionID, status, finishedAt string) error {
	if sessionID == "" {
		return fmt.Errorf("store: session id required")
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE mock_session SET status = ?, finished_at = ? WHERE id = ?
	`, status, finishedAt, sessionID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) SaveStory(ctx context.Context, story *interview.Story) error {
	if story == nil || story.ApplicationID == "" {
		return fmt.Errorf("store: story requires application_id")
	}
	if story.ID == "" {
		story.ID = uuid.NewString()
	}
	if story.EvidenceJSON == "" {
		story.EvidenceJSON = "[]"
	}
	if story.CreatedAt == "" {
		story.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO story_bank (id, application_id, category, situation, task, action, result, learning, evidence_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, story.ID, story.ApplicationID, story.Category, story.Situation, story.Task, story.Action, story.Result, story.Learning, story.EvidenceJSON, story.CreatedAt)
	return err
}

func (s *Store) GetStory(ctx context.Context, id string) (*interview.Story, error) {
	var st interview.Story
	err := s.db.QueryRowContext(ctx, `
		SELECT id, application_id, category, situation, task, action, result, learning, evidence_json, created_at
		FROM story_bank WHERE id = ?
	`, id).Scan(&st.ID, &st.ApplicationID, &st.Category, &st.Situation, &st.Task, &st.Action, &st.Result, &st.Learning, &st.EvidenceJSON, &st.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *Store) ListStories(ctx context.Context, applicationID string) ([]interview.Story, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, application_id, category, situation, task, action, result, learning, evidence_json, created_at
		FROM story_bank WHERE application_id = ? ORDER BY created_at DESC
	`, applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []interview.Story
	for rows.Next() {
		var st interview.Story
		if err := rows.Scan(&st.ID, &st.ApplicationID, &st.Category, &st.Situation, &st.Task, &st.Action, &st.Result, &st.Learning, &st.EvidenceJSON, &st.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

func (s *Store) DeleteStory(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM story_bank WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) CreateMockSession(ctx context.Context, sess *interview.MockSession, steps []interview.MockStep) error {
	if sess == nil || sess.ApplicationID == "" {
		return fmt.Errorf("store: mock session requires application_id")
	}
	if sess.ID == "" {
		sess.ID = uuid.NewString()
	}
	if sess.StartedAt == "" {
		sess.StartedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if sess.Status == "" {
		sess.Status = "active"
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO mock_session (id, application_id, type, prompt, started_at, finished_at, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, sess.ID, sess.ApplicationID, sess.Type, sess.Prompt, sess.StartedAt, sess.FinishedAt, sess.Status); err != nil {
		return err
	}
	for _, step := range steps {
		stepID := step.ID
		if stepID == "" {
			stepID = uuid.NewString()
		}
		weak := 0
		if step.Weak {
			weak = 1
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO mock_step (id, session_id, idx, name, timebox_s, output, note, weak)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, stepID, sess.ID, step.Idx, step.Name, step.TimeboxS, step.Output, step.Note, weak); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) GetMockSession(ctx context.Context, id string) (*interview.MockSession, error) {
	var sess interview.MockSession
	err := s.db.QueryRowContext(ctx, `
		SELECT id, application_id, type, prompt, started_at, finished_at, status
		FROM mock_session WHERE id = ?
	`, id).Scan(&sess.ID, &sess.ApplicationID, &sess.Type, &sess.Prompt, &sess.StartedAt, &sess.FinishedAt, &sess.Status)
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *Store) ListMockSessions(ctx context.Context, applicationID string) ([]interview.MockSession, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, application_id, type, prompt, started_at, finished_at, status
		FROM mock_session WHERE application_id = ? ORDER BY started_at DESC
	`, applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []interview.MockSession
	for rows.Next() {
		var sess interview.MockSession
		if err := rows.Scan(&sess.ID, &sess.ApplicationID, &sess.Type, &sess.Prompt, &sess.StartedAt, &sess.FinishedAt, &sess.Status); err != nil {
			return nil, err
		}
		out = append(out, sess)
	}
	return out, rows.Err()
}

func (s *Store) ListMockSteps(ctx context.Context, sessionID string) ([]interview.MockStep, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, session_id, idx, name, timebox_s, output, note, weak
		FROM mock_step WHERE session_id = ? ORDER BY idx ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []interview.MockStep
	for rows.Next() {
		var step interview.MockStep
		var weak int
		if err := rows.Scan(&step.ID, &step.SessionID, &step.Idx, &step.Name, &step.TimeboxS, &step.Output, &step.Note, &weak); err != nil {
			return nil, err
		}
		step.Weak = weak != 0
		out = append(out, step)
	}
	return out, rows.Err()
}

func (s *Store) UpdateMockStep(ctx context.Context, step *interview.MockStep) error {
	if step == nil || step.ID == "" {
		return fmt.Errorf("store: mock step id required")
	}
	weak := 0
	if step.Weak {
		weak = 1
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE mock_step SET output = ?, note = ?, weak = ? WHERE id = ?
	`, step.Output, step.Note, weak, step.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) SaveRubricScore(ctx context.Context, score *interview.RubricScore) error {
	if score == nil || score.SessionID == "" || score.Dimension == "" {
		return fmt.Errorf("store: rubric score requires session_id and dimension")
	}
	if score.ID == "" {
		score.ID = uuid.NewString()
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO rubric_score (id, session_id, dimension, score, comment)
		VALUES (?, ?, ?, ?, ?)
	`, score.ID, score.SessionID, score.Dimension, score.Score, score.Comment)
	return err
}

func (s *Store) ListRubricScores(ctx context.Context, sessionID string) ([]interview.RubricScore, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, session_id, dimension, score, comment FROM rubric_score WHERE session_id = ? ORDER BY dimension
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []interview.RubricScore
	for rows.Next() {
		var sc interview.RubricScore
		if err := rows.Scan(&sc.ID, &sc.SessionID, &sc.Dimension, &sc.Score, &sc.Comment); err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

func (s *Store) GetTopicNote(ctx context.Context, slug string) (*interview.TopicNote, error) {
	var n interview.TopicNote
	err := s.db.QueryRowContext(ctx, `
		SELECT slug, title, body, tier FROM topic_note WHERE slug = ?
	`, slug).Scan(&n.Slug, &n.Title, &n.Body, &n.Tier)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (s *Store) ListTopicNotes(ctx context.Context, tier string) ([]interview.TopicNote, error) {
	var rows *sql.Rows
	var err error
	if tier == "" {
		rows, err = s.db.QueryContext(ctx, `SELECT slug, title, body, tier FROM topic_note ORDER BY tier, slug`)
	} else {
		rows, err = s.db.QueryContext(ctx, `SELECT slug, title, body, tier FROM topic_note WHERE tier = ? ORDER BY slug`, tier)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []interview.TopicNote
	for rows.Next() {
		var n interview.TopicNote
		if err := rows.Scan(&n.Slug, &n.Title, &n.Body, &n.Tier); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *Store) GetDrillForTopic(ctx context.Context, topicSlug string) (*interview.Drill, error) {
	var d interview.Drill
	err := s.db.QueryRowContext(ctx, `
		SELECT id, topic_slug, prompt, minutes FROM drill WHERE topic_slug = ? LIMIT 1
	`, topicSlug).Scan(&d.ID, &d.TopicSlug, &d.Prompt, &d.Minutes)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
