package hunkdecision

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/pahlevikun/applydiff/internal/tailor"
)

const (
	DecisionAccept = "accept"
	DecisionReject = "reject"
	DecisionEdit   = "edit"
	DecisionSkip   = "skip"
)

// ContentHash fingerprints a hunk body for stale detection after re-tailoring.
func ContentHash(h tailor.Hunk) string {
	sum := sha256.Sum256([]byte(h.FieldID + "\x00" + h.Before + "\x00" + h.After))
	return hex.EncodeToString(sum[:])
}

// ValidDecision reports whether s is a known decision value.
func ValidDecision(s string) bool {
	switch s {
	case DecisionAccept, DecisionReject, DecisionEdit, DecisionSkip:
		return true
	default:
		return false
	}
}

// IsDecided reports whether a stored decision counts toward the build gate.
func IsDecided(decision string, text string, stale bool) bool {
	if stale || decision == "" {
		return false
	}
	switch decision {
	case DecisionAccept, DecisionReject, DecisionEdit:
		return true
	case DecisionSkip:
		return text != ""
	default:
		return false
	}
}

// Record is one persisted hunk decision.
type Record struct {
	ApplicationID string `json:"application_id"`
	HunkIndex     int    `json:"hunk_index"`
	Decision      string `json:"decision"`
	Text          string `json:"text"`
	ContentHash   string `json:"content_hash"`
	DecidedAt     string `json:"decided_at"`
}

// ApplyToHunk merges a record onto a proposal hunk (mutates h).
func ApplyToHunk(h *tailor.Hunk, rec Record, stale bool) error {
	if h == nil {
		return fmt.Errorf("hunkdecision: hunk is nil")
	}
	if stale {
		h.Status = tailor.StatusPending
		return nil
	}
	switch rec.Decision {
	case DecisionAccept:
		h.Status = tailor.StatusAccepted
	case DecisionReject:
		h.Status = tailor.StatusRejected
	case DecisionEdit:
		h.Status = tailor.StatusAccepted
		if rec.Text != "" {
			h.After = rec.Text
		}
	case DecisionSkip:
		if rec.Text == "" {
			h.Status = tailor.StatusPending
		} else {
			h.Status = tailor.StatusSkipped
		}
	default:
		h.Status = tailor.StatusPending
	}
	return nil
}
