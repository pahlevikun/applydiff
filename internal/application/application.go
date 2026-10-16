package application

import (
	"fmt"
	"time"
)

// Status values for the application lifecycle.
const (
	StatusDiscovered  = "discovered"
	StatusMatched     = "matched"
	StatusShortlisted = "shortlisted"
	StatusTailoring   = "tailoring"
	StatusReady       = "ready"
	StatusSubmitted   = "submitted"
	StatusFollowup    = "followup"
	StatusClosed      = "closed"
)

// Application tracks one job's application pipeline (separate from feed job.status).
type Application struct {
	JobID        string    `json:"job_id"`
	Status       string    `json:"status"`
	FollowUpNote string    `json:"follow_up_note"`
	CVSnapshot   string    `json:"cv_snapshot,omitempty"` // JSON frozen at export
	UpdatedAt    time.Time `json:"updated_at"`
}

// ListItem is returned by the API with job metadata joined in.
type ListItem struct {
	JobID        string    `json:"job_id"`
	Company      string    `json:"company"`
	Title        string    `json:"title"`
	URL          string    `json:"url"`
	Status       string    `json:"status"`
	FollowUpNote string    `json:"follow_up_note"`
	UpdatedAt    time.Time `json:"updated_at"`
}

var legalTransitions = map[string][]string{
	StatusDiscovered:  {StatusMatched, StatusShortlisted, StatusClosed},
	StatusMatched:     {StatusShortlisted, StatusTailoring, StatusClosed},
	StatusShortlisted: {StatusTailoring, StatusReady, StatusClosed},
	StatusTailoring:   {StatusReady, StatusClosed},
	StatusReady:       {StatusSubmitted, StatusClosed},
	StatusSubmitted:   {StatusFollowup, StatusClosed},
	StatusFollowup:    {StatusClosed},
	StatusClosed:      {},
}

// CanTransition reports whether moving from one status to another is allowed.
func CanTransition(from, to string) bool {
	if from == to {
		return true
	}
	next, ok := legalTransitions[from]
	if !ok {
		return false
	}
	for _, s := range next {
		if s == to {
			return true
		}
	}
	return false
}

// Transition validates and returns the new status or an error.
func Transition(from, to string) (string, error) {
	if to == "" {
		return "", fmt.Errorf("application: target status required")
	}
	if !validStatus(to) {
		return "", fmt.Errorf("application: unknown status %q", to)
	}
	if from == "" {
		from = StatusDiscovered
	}
	if !validStatus(from) {
		return "", fmt.Errorf("application: unknown status %q", from)
	}
	if !CanTransition(from, to) {
		return "", fmt.Errorf("application: illegal transition %s -> %s", from, to)
	}
	return to, nil
}

func validStatus(s string) bool {
	switch s {
	case StatusDiscovered, StatusMatched, StatusShortlisted, StatusTailoring,
		StatusReady, StatusSubmitted, StatusFollowup, StatusClosed:
		return true
	default:
		return false
	}
}
