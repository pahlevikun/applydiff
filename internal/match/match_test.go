package match

import (
	"strings"
	"testing"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/job"
)

func TestScoreDimensionsTotalRange(t *testing.T) {
	doc := fixtureCV()
	j := job.Job{
		ID:          "job-1",
		Company:     "Acme Corp",
		Title:       "Senior Backend Engineer",
		URL:         "https://boards.greenhouse.io/acme/jobs/123",
		Source:      "greenhouse",
		Remote:      "yes",
		Description: fixtureJD(),
		SalaryPost:  "120000-150000",
		SalaryATS:   "120000-150000",
	}
	res := Evaluate(Input{Job: j, CV: doc})
	if res.Total < 0 || res.Total > 25 {
		t.Fatalf("total out of range: %d", res.Total)
	}
	if res.Dimensions.Total() != res.Total {
		t.Fatalf("dimension sum %d != total %d", res.Dimensions.Total(), res.Total)
	}
	for name, v := range map[string]int{
		"skills": res.Dimensions.Skills,
		"seniority": res.Dimensions.Seniority,
		"domain": res.Dimensions.Domain,
		"logistics": res.Dimensions.Logistics,
		"motivation": res.Dimensions.Motivation,
	} {
		if v < 0 || v > 5 {
			t.Fatalf("%s score %d out of 0-5", name, v)
		}
	}
	if res.Dimensions.Skills < 3 {
		t.Fatalf("expected strong skills match, got %d", res.Dimensions.Skills)
	}
	if res.Dimensions.Seniority != 5 {
		t.Fatalf("expected seniority 5 for senior title match, got %d", res.Dimensions.Seniority)
	}
}

func TestRequirementLabels(t *testing.T) {
	doc := fixtureCV()
	reqs := ExtractRequirements(fixtureJD(), doc)
	if len(reqs) == 0 {
		t.Fatal("expected requirements from fixture JD")
	}
	var hasMatch, hasGap bool
	for _, r := range reqs {
		switch r.Status {
		case StatusMatch:
			hasMatch = true
			if !strings.Contains(strings.ToLower(r.Phrase), "go") {
				t.Logf("match phrase: %s", r.Phrase)
			}
		case StatusGap:
			hasGap = true
			if r.CVFieldID != "" {
				t.Fatalf("gap should not have cv field: %+v", r)
			}
		}
	}
	if !hasMatch {
		t.Fatal("expected at least one match requirement")
	}
	gapPhrase := "- Required: 5+ years of COBOL mainframe experience"
	gap := classifyRequirement(gapPhrase, doc)
	if gap.Status != StatusGap {
		t.Fatalf("expected genuine gap, got %s", gap.Status)
	}
	if !hasGap {
		for _, r := range reqs {
			if strings.Contains(strings.ToLower(r.Phrase), "cobol") {
				hasGap = true
			}
		}
	}
	jdWithGap := fixtureJD() + "\n" + gapPhrase + "\n"
	if !hasGap {
		reqs2 := ExtractRequirements(jdWithGap, doc)
		for _, r := range reqs2 {
			if strings.Contains(strings.ToLower(r.Phrase), "cobol") && r.Status == StatusGap {
				hasGap = true
			}
		}
	}
	if !hasGap {
		t.Fatal("expected a gap requirement (COBOL)")
	}
}

func TestLegitimacyATSAndSalary(t *testing.T) {
	j := job.Job{
		Company:    "Stripe",
		URL:        "https://boards.greenhouse.io/stripe/jobs/123",
		SalaryPost: "$120,000 - $150,000 USD",
		SalaryATS:  "$120,000 - $150,000 USD",
	}
	leg := CheckLegitimacy(j)
	if !leg.OfficialATSURL {
		t.Fatal("expected official ATS URL")
	}
	if !leg.CompanyVerified {
		t.Fatal("expected company verified")
	}
	if !leg.SalaryConsistent {
		t.Fatal("expected salary consistent")
	}

	j.SalaryATS = "$90,000 - $110,000"
	leg = CheckLegitimacy(j)
	if leg.SalaryConsistent {
		t.Fatal("expected salary mismatch flag")
	}
	if len(leg.Flags) == 0 || leg.Flags[0] != "salary_mismatch" {
		t.Fatalf("expected salary_mismatch flag, got %v", leg.Flags)
	}
}

func fixtureJD() string {
	return `Senior Backend Engineer

Requirements:
- Required: 5+ years experience with Go
- Strong PostgreSQL and distributed systems background
- Experience building APIs at scale
- Must have Kubernetes experience
`
}

func fixtureCV() *cv.Document {
	end := "2022-12"
	return &cv.Document{
		Person: "test-person",
		Contact: cv.Contact{
			Name:         "Test User",
			Headline:     "Senior Software Engineer",
			Email:        "test@example.com",
			TimezoneLine: "GMT+7 · available for EU overlap",
		},
		Summary: cv.Summary{ID: "summary", Text: "Backend engineer focused on APIs and distributed systems at scale."},
		Skills: []cv.SkillGroup{
			{ID: "skills.backend", Label: "Backend", Items: []string{"Go", "PostgreSQL", "Kubernetes"}},
		},
		Experience: []cv.Experience{
			{
				ID: "exp.1", Company: "Co", Title: "Senior Engineer", Start: "2020-01", End: &end,
				Bullets: []cv.Bullet{{ID: "exp.1.b1", Text: "Built high-scale APIs"}},
			},
		},
		Education:    []cv.Education{{ID: "edu.1", School: "Uni", Degree: "BS", Year: "2020"}},
		Achievements: []cv.Achievement{{ID: "ach.1", Text: "Shipped product impact"}},
	}
}
