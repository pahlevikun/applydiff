package extract

import (
	"regexp"
	"strings"

	"github.com/pahlevikun/applydiff/internal/job"
)

var (
	reSalaryUSD = regexp.MustCompile(`(?i)\$\s*([\d,]+)(?:\s*k)?(?:\s*[-–]\s*\$?\s*[\d,]+(?:\s*k)?)?`)
	reSalaryIDR = regexp.MustCompile(`(?i)(?:IDR|Rp\.?)\s*([\d.,]+)\s*(?:jt|juta|m)?`)
	reOverlap   = regexp.MustCompile(`(?i)(\d+)\+?\s*hours?\s+overlap\s+with\s+([A-Z]{2,4})`)
	re4Day      = regexp.MustCompile(`(?i)(4[- ]day|four[- ]day)\s+week`)
	re45Day     = regexp.MustCompile(`(?i)4\.5[- ]day`)
)

// FromText populates filterable fields on j from title + description (heuristic, no LLM).
func FromText(j *job.Job) {
	if j == nil {
		return
	}
	text := j.Description
	if text == "" {
		text = j.Title
	}
	lower := strings.ToLower(text + " " + j.Title)

	applyArrangement(j, lower)
	applyWeeklyDays(j, lower)
	j.AsyncFriendly = containsAny(lower, "async-first", "async first", "flexible hours", "work async")
	applyTimezone(j, text)
	applyRegionCountry(j, lower)
	applyEmployment(j, lower)
	j.Seniority = detectSeniority(lower)
	applySalary(j, text)
	j.VisaSponsorship = containsAny(lower, "visa sponsorship", "sponsor visa", "work visa sponsorship")
	j.Relocation = containsAny(lower, "relocation assistance", "relocation package", "relocation offered")
	j.EORFriendly = containsAny(lower, "eor", "employer of record", "deel", "remote.com")
	j.Archetype = detectArchetype(lower)
	j.Keywords = detectKeywords(lower)
	j.Benefits = detectBenefits(lower)
	j.Equity = containsAny(lower, "equity", "stock options", "rsu")

	syncLegacyRemote(j)
}

func applyArrangement(j *job.Job, lower string) {
	if containsAny(lower, "fully remote", "100% remote", "remote-first", "work from anywhere", "fully-remote") ||
		(strings.Contains(lower, "remote") && !strings.Contains(lower, "hybrid") && !strings.Contains(lower, "on-site") && !strings.Contains(lower, "onsite")) {
		j.Arrangement = "remote"
		return
	}
	if strings.Contains(lower, "hybrid") {
		j.Arrangement = "hybrid"
		for _, n := range []string{"1 day", "2 day", "3 day", "4 day"} {
			if strings.Contains(lower, n+" on-site") || strings.Contains(lower, n+" onsite") {
				j.DaysOnsite = int(n[0] - '0')
				break
			}
		}
		return
	}
	if strings.Contains(lower, "on-site") || strings.Contains(lower, "onsite") || strings.Contains(lower, "in-office") {
		j.Arrangement = "onsite"
	}
}

func applyWeeklyDays(j *job.Job, lower string) {
	if re4Day.MatchString(lower) || strings.Contains(lower, "compressed week") ||
		(strings.Contains(lower, "four-day") && strings.Contains(lower, "week")) {
		j.WeeklyDays = 4
		return
	}
	if re45Day.MatchString(lower) {
		j.WeeklyDays = 4.5
	}
}

func applyTimezone(j *job.Job, text string) {
	if m := reOverlap.FindStringSubmatch(text); len(m) == 3 {
		n := 0
		for _, c := range m[1] {
			n = n*10 + int(c-'0')
		}
		j.TimezoneOverlapHours = n
	}
}

func applyRegionCountry(j *job.Job, lower string) {
	if containsAny(lower, "southeast asia", " sea ", "singapore", "indonesia", "malaysia", "thailand", "vietnam", "philippines") {
		if j.Region == "" {
			j.Region = "SEA"
		}
	}
	if containsAny(lower, "european union", " eu ", "europe", "berlin", "london", "amsterdam") {
		if j.Region == "" {
			j.Region = "EU"
		}
	}
	if containsAny(lower, "worldwide", "anywhere", "global") {
		if j.Region == "" {
			j.Region = "global"
		}
	}
	if strings.Contains(lower, "indonesia") {
		j.Country = "ID"
	}
	if strings.Contains(lower, "singapore") {
		j.Country = "SG"
	}
}

func applyEmployment(j *job.Job, lower string) {
	switch {
	case strings.Contains(lower, "full-time") || strings.Contains(lower, "full time"):
		j.EmploymentType = "full-time"
	case strings.Contains(lower, "part-time"):
		j.EmploymentType = "part-time"
	case strings.Contains(lower, "contract-to-hire"):
		j.EmploymentType = "contract-to-hire"
	case strings.Contains(lower, "contract"):
		j.EmploymentType = "contract"
	case strings.Contains(lower, "freelance"):
		j.EmploymentType = "freelance"
	case strings.Contains(lower, "internship"):
		j.EmploymentType = "internship"
	}
}

func detectSeniority(lower string) string {
	order := []string{"director", "principal", "staff", "lead", "senior", "mid", "junior", "intern"}
	for _, s := range order {
		if strings.Contains(lower, s) {
			return s
		}
	}
	return ""
}

func applySalary(j *job.Job, text string) {
	if m := reSalaryUSD.FindStringSubmatch(text); len(m) == 2 {
		j.SalaryDisclosed = true
		j.SalaryCurrency = "USD"
		j.SalaryMin = parseAmountUSD(m[1])
		return
	}
	if m := reSalaryIDR.FindStringSubmatch(text); len(m) == 2 {
		j.SalaryDisclosed = true
		j.SalaryCurrency = "IDR"
		j.SalaryMin = parseAmountIDR(m[1])
	}
}

func parseAmountUSD(s string) int {
	s = strings.ReplaceAll(s, ",", "")
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	if n > 0 && n < 1000 {
		n *= 1000
	}
	return n
}

func parseAmountIDR(s string) int {
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	if strings.Contains(strings.ToLower(s), "jt") || n < 1000 {
		n *= 1_000_000
	}
	return n
}

func detectArchetype(lower string) string {
	rules := []struct {
		val string
		kws []string
	}{
		{"backend", []string{"backend", "back-end"}},
		{"frontend", []string{"frontend", "front-end"}},
		{"fullstack", []string{"full stack", "fullstack"}},
		{"platform", []string{"platform engineer"}},
		{"devops", []string{"devops", "sre", "site reliability"}},
		{"data", []string{"data engineer"}},
		{"ml", []string{"machine learning", "ml engineer", "ai engineer"}},
		{"mobile", []string{"mobile", "ios", "android"}},
		{"security", []string{"security engineer"}},
		{"qa", []string{"qa engineer", "quality assurance"}},
	}
	for _, r := range rules {
		for _, kw := range r.kws {
			if strings.Contains(lower, kw) {
				return r.val
			}
		}
	}
	return ""
}

func detectKeywords(lower string) []string {
	candidates := []string{"go", "golang", "kubernetes", "postgres", "aws", "distributed systems"}
	var out []string
	for _, c := range candidates {
		if strings.Contains(lower, c) {
			out = append(out, c)
		}
	}
	return out
}

func detectBenefits(lower string) []string {
	var out []string
	if re4Day.MatchString(lower) {
		out = append(out, "4-day week")
	}
	if strings.Contains(lower, "visa sponsorship") {
		out = append(out, "visa sponsorship")
	}
	if strings.Contains(lower, "equity") {
		out = append(out, "equity")
	}
	if strings.Contains(lower, "learning budget") {
		out = append(out, "learning budget")
	}
	if strings.Contains(lower, "health insurance") {
		out = append(out, "health insurance")
	}
	return out
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func syncLegacyRemote(j *job.Job) {
	if j.Arrangement == "remote" {
		j.Remote = "yes"
	}
}
