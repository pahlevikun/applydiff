package match

import (
	"net/url"
	"strings"

	"github.com/pahlevikun/applydiff/internal/job"
)

var atsHosts = []string{
	"greenhouse.io",
	"boards.greenhouse.io",
	"jobs.lever.co",
	"lever.co",
	"jobs.ashbyhq.com",
	"ashbyhq.com",
	"apply.workable.com",
	"workable.com",
}

// CheckLegitimacy applies lightweight heuristics for posting trust signals.
func CheckLegitimacy(j job.Job) Legitimacy {
	leg := Legitimacy{
		CompanyVerified:  companyVerified(j.Company),
		OfficialATSURL:   officialATSURL(j.URL),
		SalaryConsistent: true,
	}
	post := normalizeSalary(j.SalaryPost)
	ats := normalizeSalary(j.SalaryATS)
	if post != "" && ats != "" && post != ats {
		leg.SalaryConsistent = false
		leg.Flags = append(leg.Flags, "salary_mismatch")
	}
	return leg
}

func companyVerified(company string) bool {
	c := strings.TrimSpace(company)
	if len(c) < 2 {
		return false
	}
	lower := strings.ToLower(c)
	if lower == "unknown" || lower == "tbd" || lower == "n/a" {
		return false
	}
	return true
}

func officialATSURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	host := strings.ToLower(u.Host)
	for _, h := range atsHosts {
		if host == h || strings.HasSuffix(host, "."+h) {
			return true
		}
	}
	return false
}

func normalizeSalary(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, "usd", "")
	return s
}
