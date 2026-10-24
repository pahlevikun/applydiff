package store

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

const KindNeedsDecisionDigest = "needs_decision_digest"

// NotificationPayload for needs_decision_digest.
type NotificationPayload struct {
	RunID  string   `json:"run_id"`
	Count  int      `json:"count"`
	Titles []string `json:"titles,omitempty"`
	Link   string   `json:"link,omitempty"`
}

// Notification is one outbound message attempt.
type Notification struct {
	ID          string
	Kind        string
	PayloadJSON string
	CreatedAt   time.Time
	SentAt      *time.Time
	Error       string
}

// TryInsertNotification records a pending notification. Returns false if (kind, run_id) already exists.
func (s *Store) TryInsertNotification(ctx context.Context, kind string, payload NotificationPayload) (string, bool, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", false, err
	}
	id := uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO notifications (id, kind, payload_json, created_at, sent_at, error)
		VALUES (?, ?, ?, ?, NULL, '')
	`, id, kind, string(raw), now)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return "", false, nil
		}
		return "", false, err
	}
	return id, true, nil
}

// MarkNotificationSent sets sent_at after a successful send.
func (s *Store) MarkNotificationSent(ctx context.Context, id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `UPDATE notifications SET sent_at = ?, error = '' WHERE id = ?`, now, id)
	return err
}

// MarkNotificationFailed records the send error.
func (s *Store) MarkNotificationFailed(ctx context.Context, id string, sendErr error) error {
	msg := ""
	if sendErr != nil {
		msg = sendErr.Error()
	}
	_, err := s.db.ExecContext(ctx, `UPDATE notifications SET error = ? WHERE id = ?`, msg, id)
	return err
}

// CountNotifications counts rows for a kind (empty kind = all).
func (s *Store) CountNotifications(ctx context.Context, kind string) (int, error) {
	var n int
	if kind == "" {
		err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications`).Scan(&n)
		return n, err
	}
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications WHERE kind = ?`, kind).Scan(&n)
	return n, err
}
