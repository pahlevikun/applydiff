package extract

import (
	"testing"

	"github.com/pahlevikun/applydiff/internal/job"
)

const fixtureJD = `
Senior Backend Engineer (Remote, SEA)

We are a fully remote team hiring from Southeast Asia. This is a four-day work week.
Salary: $120,000 - $150,000 USD per year.

Requirements:
- 4+ hours overlap with EST
- Go and distributed systems

Visa sponsorship available for exceptional candidates.
We work async-first with flexible hours.
`

func TestExtractFixtureJD(t *testing.T) {
	j := &job.Job{
		Title:       "Senior Backend Engineer",
		Description: fixtureJD,
	}
	FromText(j)

	if j.Arrangement != "remote" {
		t.Fatalf("arrangement %q", j.Arrangement)
	}
	if j.WeeklyDays != 4 {
		t.Fatalf("weekly_days %v", j.WeeklyDays)
	}
	if j.Region != "SEA" {
		t.Fatalf("region %q", j.Region)
	}
	if !j.SalaryDisclosed || j.SalaryCurrency != "USD" || j.SalaryMin < 100000 {
		t.Fatalf("salary disclosed=%v currency=%s min=%d", j.SalaryDisclosed, j.SalaryCurrency, j.SalaryMin)
	}
	if !j.VisaSponsorship {
		t.Fatal("expected visa sponsorship")
	}
	if j.TimezoneOverlapHours < 4 {
		t.Fatalf("overlap hours %d", j.TimezoneOverlapHours)
	}
	if !j.AsyncFriendly {
		t.Fatal("expected async-friendly")
	}
}
