package job

import (
	"testing"
	"time"
)

func TestRestamped(t *testing.T) {
	now := time.Now().UTC()
	first := now.Add(-40 * 24 * time.Hour).Format(time.RFC3339)
	recent := now.Add(-2 * 24 * time.Hour).Format(time.RFC3339)
	postedOld := now.Format(time.RFC3339)

	if !Restamped(postedOld, first) {
		t.Fatal("expected restamped when 40 days apart")
	}
	if Restamped(postedOld, recent) {
		t.Fatal("expected not restamped when 2 days apart")
	}
	if Restamped("not-a-date", first) {
		t.Fatal("expected false for unparsable posted_at")
	}
	if Restamped(postedOld, "not-a-date") {
		t.Fatal("expected false for unparsable first_seen")
	}
}
