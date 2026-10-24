package store

import (
	"context"
	"testing"
)

func TestNotificationDedupeByKindAndRunID(t *testing.T) {
	ctx := context.Background()
	st, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	payload := NotificationPayload{RunID: "run-1", Count: 3}
	id1, ok1, err := st.TryInsertNotification(ctx, KindNeedsDecisionDigest, payload)
	if err != nil || !ok1 || id1 == "" {
		t.Fatalf("first insert: id=%q ok=%v err=%v", id1, ok1, err)
	}
	_, ok2, err := st.TryInsertNotification(ctx, KindNeedsDecisionDigest, payload)
	if err != nil {
		t.Fatal(err)
	}
	if ok2 {
		t.Fatal("expected duplicate insert to be ignored")
	}
	n, err := st.CountNotifications(ctx, KindNeedsDecisionDigest)
	if err != nil || n != 1 {
		t.Fatalf("count=%d err=%v", n, err)
	}
}
