package risk

import (
	"testing"

	"github.com/pahlevikun/applydiff/internal/job"
)

func TestAssessRules(t *testing.T) {
	peerDup := map[string]bool{
		DescriptionHash("We build payments APIs for global merchants. " + stringsRepeat("detail ", 20)): true,
	}

	cases := []struct {
		name     string
		j        job.Job
		ctx      Context
		wantCode string
		wantSev  string
		wantZero bool
	}{
		{
			name: "email-domain-mismatch",
			j: job.Job{
				Company:     "Acme Corp",
				URL:         "https://acme.com/jobs/1",
				Description: "Apply at hr@gmail.com for this backend role at Acme.",
			},
			wantCode: codeEmailDomainMismatch,
			wantSev:  SeverityHard,
		},
		{
			name: "pay-to-work",
			j: job.Job{
				Company:     "Acme",
				URL:         "https://acme.com/j/1",
				Description: "Candidates must pay a training fee before onboarding.",
			},
			wantCode: codePayToWork,
			wantSev:  SeverityHard,
		},
		{
			name: "id-before-interview",
			j: job.Job{
				Company:     "Acme",
				URL:         "https://acme.com/j/1",
				Description: "Send your passport and bank account before the interview.",
			},
			wantCode: codeIDBeforeInterview,
			wantSev:  SeverityHard,
		},
		{
			name: "clone-and-run",
			j: job.Job{
				Company:     "Acme",
				URL:         "https://acme.com/j/1",
				Description: "Please clone this repo and run npm install locally as your technical test.",
			},
			wantCode: codeCloneAndRun,
			wantSev:  SeverityHard,
		},
		{
			name: "contact-without-domain",
			j: job.Job{
				Company:     "Mystery LLC",
				Description: "Reach us on WhatsApp +1-555-0100 only.",
			},
			wantCode: codeContactWithoutDomain,
			wantSev:  SeverityHard,
		},
		{
			name: "vague-location",
			j: job.Job{
				Company:     "Acme",
				URL:         "https://acme.com/j/1",
				Description: "Fully remote worldwide role with async collaboration.",
			},
			wantCode: codeVagueLocation,
			wantSev:  SeveritySoft,
		},
		{
			name: "duplicate-jd",
			j: job.Job{
				ID:          "job-b",
				Company:     "Other",
				URL:         "https://other.com/1",
				Description: "We build payments APIs for global merchants. " + stringsRepeat("detail ", 20),
			},
			ctx:      Context{PeerDescriptionHashes: peerDup},
			wantCode: codeDuplicateJD,
			wantSev:  SeveritySoft,
		},
		{
			name: "salary-outlier",
			j: job.Job{
				Company:         "Acme",
				URL:             "https://acme.com/j/1",
				Title:           "Junior Engineer",
				Seniority:       "junior",
				SalaryDisclosed: true,
				SalaryMin:       400_000,
				SalaryCurrency:  "USD",
				Description:     "Junior backend engineer building APIs.",
			},
			wantCode: codeSalaryOutlier,
			wantSev:  SeveritySoft,
		},
		{
			name: "unnamed-company",
			j: job.Job{
				Company:     "Confidential",
				URL:         "https://example.com/j/1",
				Description: "Senior engineer role.",
			},
			wantCode: codeUnnamedCompany,
			wantSev:  SeveritySoft,
		},
		{
			name: "grounded-clean",
			j: job.Job{
				Company:     "Acme Payments",
				URL:         "https://boards.greenhouse.io/acme/jobs/123",
				Title:       "Senior Backend Engineer",
				Seniority:   "senior",
				Region:      "EU",
				Description: "Acme Payments is hiring a senior backend engineer in our Berlin office hybrid 2 days/week. Apply via our greenhouse link. Salary $140k-$180k USD.",
				SalaryDisclosed: true,
				SalaryMin:       140_000,
				SalaryCurrency:  "USD",
			},
			wantZero: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := Assess(tc.j, tc.ctx)
			if tc.wantZero {
				if len(a.Flags) != 0 || a.Score != 0 {
					t.Fatalf("expected clean score, got score=%d flags=%v", a.Score, a.Flags)
				}
				return
			}
			var found bool
			for _, f := range a.Flags {
				if f.Code == tc.wantCode {
					found = true
					if f.Severity != tc.wantSev {
						t.Fatalf("severity %q want %q", f.Severity, tc.wantSev)
					}
				}
			}
			if !found {
				t.Fatalf("missing flag %s in %v", tc.wantCode, a.Flags)
			}
		})
	}
}

func stringsRepeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
