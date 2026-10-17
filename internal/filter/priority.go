package filter

import (
	"math"
	"time"

	"github.com/pahlevikun/applydiff/internal/job"
)

// Priority computes freshness × match × (1/competition).
// Applicant count 0 means competition factor is neutral (1).
func Priority(j job.Job, matchScore int) float64 {
	fresh := freshnessFactor(j.PostedAt)
	match := math.Max(1, float64(matchScore+1))
	comp := 1.0
	if j.ApplicantCount > 0 {
		comp = 1.0 / float64(j.ApplicantCount)
	}
	return fresh * match * comp
}

func freshnessFactor(postedAt string) float64 {
	t, ok := parsePosted(postedAt)
	if !ok {
		return 0.5
	}
	hours := time.Since(t).Hours()
	if hours <= 0 {
		return 1
	}
	return 1.0 / (1.0 + hours/24.0)
}
