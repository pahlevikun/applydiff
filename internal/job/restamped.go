package job

import (
	"strings"
	"time"
)

const restampThreshold = 7 * 24 * time.Hour

// Restamped is true when posted_at is more than 7 days after first_seen (both parse as times).
func Restamped(postedAt, firstSeen string) bool {
	postedAt = strings.TrimSpace(postedAt)
	firstSeen = strings.TrimSpace(firstSeen)
	if postedAt == "" || firstSeen == "" {
		return false
	}
	p, err := parseJobTime(postedAt)
	if err != nil {
		return false
	}
	f, err := parseJobTime(firstSeen)
	if err != nil {
		return false
	}
	return p.Sub(f) > restampThreshold
}

func parseJobTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.UTC(), nil
	}
	return time.Parse("2006-01-02", s)
}
