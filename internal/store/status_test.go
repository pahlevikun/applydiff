package store

import (
	"context"
	"testing"
	"time"
)

func TestLoadSystemStatusFromSourceScans(t *testing.T) {
	st, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	at := "2026-10-20T12:00:00Z"
	if err := st.SaveSourceScanResult(context.Background(), "stub", SourceScanResult{
		At: at, Fetched: 1, Stored: 1,
	}); err != nil {
		t.Fatal(err)
	}
	got, err := st.LoadSystemStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.LastScanAt == nil {
		t.Fatal("expected last_scan_at")
	}
	want, _ := time.Parse(time.RFC3339, at)
	if !got.LastScanAt.Equal(want) {
		t.Fatalf("last scan %v want %v", got.LastScanAt, want)
	}
}
