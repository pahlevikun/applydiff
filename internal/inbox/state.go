package inbox

import (
	"strings"

	"github.com/pahlevikun/applydiff/internal/application"
	"github.com/pahlevikun/applydiff/internal/store"
)

// State is a queue bucket for decision workflow.
type State string

const (
	StateNeedsDecision State = "needs-decision"
	StateInProgress    State = "in-progress"
	StateSent          State = "sent"
	StateInterview     State = "interview"
)

// MatchMinScore is the same threshold used by runtime.ComputeMatch.
const MatchMinScore = 8

// Classify maps one joined store row to a queue state and human reason.
// When multiple states could apply, the earlier state wins:
// needs-decision > in-progress > sent > interview.
func Classify(row store.InboxRow) (State, string) {
	if st, reason, ok := classifyNeedsDecision(row); ok {
		return st, reason
	}
	if st, reason, ok := classifyInProgress(row); ok {
		return st, reason
	}
	if st, reason, ok := classifySent(row); ok {
		return st, reason
	}
	if st, reason, ok := classifyInterview(row); ok {
		return st, reason
	}
	return "", ""
}

func jobActive(row store.InboxRow) bool {
	if row.Job.Status == "dead" || row.Job.Status == "discarded" {
		return false
	}
	if row.Job.Status == "skipped" || strings.TrimSpace(row.Job.FilterSkipReason) != "" {
		return false
	}
	return true
}

// needs-decision: passed filters, user has not committed to the pipeline, and there is
// something to review (match at/above threshold or an existing draft).
func classifyNeedsDecision(row store.InboxRow) (State, string, bool) {
	if !jobActive(row) {
		return "", "", false
	}
	if row.AppStatus == application.StatusSubmitted || row.AppStatus == application.StatusFollowup ||
		row.AppStatus == application.StatusClosed {
		return "", "", false
	}
	if row.AppStatus != "" && row.AppStatus != application.StatusDiscovered &&
		row.AppStatus != application.StatusMatched && row.AppStatus != application.StatusShortlisted {
		return "", "", false
	}
	hasReviewable := row.HasProposal || row.MatchTotal >= MatchMinScore
	if !hasReviewable {
		return "", "", false
	}
	if row.HasProposal {
		return StateNeedsDecision, "draft ready for review", true
	}
	return StateNeedsDecision, "match score ready for decision", true
}

// in-progress: pipeline started but tailoring/diff not fully approved or not built yet.
func classifyInProgress(row store.InboxRow) (State, string, bool) {
	if !jobActive(row) {
		return "", "", false
	}
	switch row.AppStatus {
	case application.StatusTailoring:
		if row.PendingHunks > 0 {
			return StateInProgress, "hunks awaiting approval", true
		}
		return StateInProgress, "tailoring in progress", true
	case application.StatusReady:
		return StateInProgress, "ready to build or send", true
	case application.StatusShortlisted:
		if row.HasProposal && row.PendingHunks > 0 {
			return StateInProgress, "shortlisted with pending hunks", true
		}
	}
	if row.HasProposal && row.PendingHunks > 0 && row.AppStatus != application.StatusSubmitted {
		return StateInProgress, "diff not fully decided", true
	}
	return "", "", false
}

func classifySent(row store.InboxRow) (State, string, bool) {
	if row.AppStatus == application.StatusSubmitted {
		return StateSent, "applied, awaiting reply", true
	}
	return "", "", false
}

func classifyInterview(row store.InboxRow) (State, string, bool) {
	if row.AppStatus == application.StatusFollowup {
		return StateInterview, "reply received; interview prep", true
	}
	return "", "", false
}
