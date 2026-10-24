package notify

import (
	"context"
	"fmt"
	"strings"

	"github.com/pahlevikun/applydiff/internal/inbox"
	"github.com/pahlevikun/applydiff/internal/store"
)

// SendNeedsDecisionDigest sends at most one digest per run when needs-decision count grew.
func SendNeedsDecisionDigest(ctx context.Context, st *store.Store, n Notifier, cfg Config, runID string, beforeCount int) error {
	if st == nil || n == nil || runID == "" {
		return nil
	}
	after, titles, err := needsDecisionSnapshot(ctx, st)
	if err != nil {
		return err
	}
	if after <= beforeCount {
		return nil
	}
	link := cfg.InboxLink()
	payload := store.NotificationPayload{
		RunID:  runID,
		Count:  after,
		Titles: titles,
		Link:   link,
	}
	id, inserted, err := st.TryInsertNotification(ctx, store.KindNeedsDecisionDigest, payload)
	if err != nil || !inserted {
		return err
	}
	msg := buildDigestMessage(after, titles, link)
	sendErr := n.Send(ctx, msg)
	if sendErr != nil {
		_ = st.MarkNotificationFailed(ctx, id, sendErr)
		return sendErr
	}
	return st.MarkNotificationSent(ctx, id)
}

func needsDecisionSnapshot(ctx context.Context, st *store.Store) (int, []string, error) {
	res, err := inbox.List(ctx, st, inbox.ListOptions{State: string(inbox.StateNeedsDecision), Limit: 5})
	if err != nil {
		return 0, nil, err
	}
	var titles []string
	for _, item := range res.Data {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = item.JobKey
		}
		titles = append(titles, title)
	}
	return res.Meta.Counts.NeedsDecision, titles, nil
}

func buildDigestMessage(count int, titles []string, link string) Message {
	title := fmt.Sprintf("%d applications need your decision", count)
	var body strings.Builder
	for _, t := range titles {
		body.WriteString("• ")
		body.WriteString(t)
		body.WriteString("\n")
	}
	return Message{
		Kind:  store.KindNeedsDecisionDigest,
		Title: title,
		Body:  strings.TrimRight(body.String(), "\n"),
		Link:  link,
	}
}

// NeedsDecisionCount returns the current needs-decision queue size.
func NeedsDecisionCount(ctx context.Context, st *store.Store) (int, error) {
	res, err := inbox.List(ctx, st, inbox.ListOptions{})
	if err != nil {
		return 0, err
	}
	return res.Meta.Counts.NeedsDecision, nil
}
