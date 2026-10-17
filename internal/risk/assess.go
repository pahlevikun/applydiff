package risk

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/pahlevikun/applydiff/internal/job"
)

const (
	codeEmailDomainMismatch   = "email-domain-mismatch"
	codePayToWork             = "pay-to-work"
	codeIDBeforeInterview     = "id-before-interview"
	codeCloneAndRun           = "clone-and-run"
	codeContactWithoutDomain  = "contact-without-domain"
	codeVagueLocation         = "vague-location"
	codeDuplicateJD           = "duplicate-jd"
	codeSalaryOutlier         = "salary-outlier"
	codeUnnamedCompany        = "unnamed-company"
)

var (
	reEmail       = regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@([a-z0-9.\-]+\.[a-z]{2,})`)
	reURLHost     = regexp.MustCompile(`(?i)https?://([a-z0-9.\-]+)`)
	reWhatsApp    = regexp.MustCompile(`(?i)(wa\.me/|whatsapp|whats\s*app)`)
	reTelegram    = regexp.MustCompile(`(?i)(t\.me/|telegram)`)
	reCloneRun    = regexp.MustCompile(`(?i)(clone\s+(this\s+)?repo|git\s+clone|clone\s+the\s+repository).{0,120}(run|execute|npm\s+start|make\s+run)`)
	rePayToWork   = regexp.MustCompile(`(?i)(pay\s+(a\s+)?fee|training\s+fee|equipment\s+purchase|security\s+deposit|send\s+\$|wire\s+transfer|upfront\s+payment|pay\s+for\s+(training|equipment|kit))`)
	reIDBeforeInt = regexp.MustCompile(`(?i)(before\s+(the\s+)?interview|prior\s+to\s+interview).{0,80}(id\s+card|passport|bank\s+account|selfie|ssn|social\s+security)`)
	reIDBeforeInt2 = regexp.MustCompile(`(?i)(send\s+(us\s+)?(your\s+)?(id\s+card|passport|bank\s+account|selfie)).{0,80}(before|prior)`)
)

var freeMailDomains = map[string]bool{
	"gmail.com": true, "googlemail.com": true, "yahoo.com": true, "hotmail.com": true,
	"outlook.com": true, "icloud.com": true, "protonmail.com": true, "mail.com": true,
}

// Assess runs deterministic rules on j. Pass ctx for duplicate-jd peer hashes.
func Assess(j job.Job, ctx Context) Assessment {
	text := strings.ToLower(j.Description + "\n" + j.Title)
	var flags []Flag

	if f := checkEmailDomainMismatch(j, text); f != nil {
		flags = append(flags, *f)
	}
	if f := checkPayToWork(text); f != nil {
		flags = append(flags, *f)
	}
	if f := checkIDBeforeInterview(text); f != nil {
		flags = append(flags, *f)
	}
	if f := checkCloneAndRun(text); f != nil {
		flags = append(flags, *f)
	}
	if f := checkContactWithoutDomain(j, text); f != nil {
		flags = append(flags, *f)
	}
	if f := checkVagueLocation(j, text); f != nil {
		flags = append(flags, *f)
	}
	if f := checkDuplicateJD(j, ctx); f != nil {
		flags = append(flags, *f)
	}
	if f := checkSalaryOutlier(j); f != nil {
		flags = append(flags, *f)
	}
	if f := checkUnnamedCompany(j); f != nil {
		flags = append(flags, *f)
	}

	score := scoreFlags(flags)
	return Assessment{Score: score, Flags: flags}
}

func scoreFlags(flags []Flag) int {
	if len(flags) == 0 {
		return 0
	}
	hard := 0
	soft := 0
	for _, f := range flags {
		if f.Severity == SeverityHard {
			hard++
		} else {
			soft++
		}
	}
	score := hard*35 + soft*12
	if hard > 0 {
		if score < 70 {
			score = 70
		}
	}
	if score > 100 {
		score = 100
	}
	return score
}

func checkEmailDomainMismatch(j job.Job, text string) *Flag {
	emails := reEmail.FindAllStringSubmatch(text, -1)
	if len(emails) == 0 {
		return nil
	}
	companyDomain := companyDomain(j)
	if companyDomain == "" {
		return nil
	}
	for _, m := range emails {
		domain := strings.ToLower(m[1])
		if freeMailDomains[domain] {
			return &Flag{
				Code:     codeEmailDomainMismatch,
				Severity: SeverityHard,
				Message:  "contact email uses a personal/free domain, not the company domain",
			}
		}
		if domain != companyDomain && !strings.HasSuffix(domain, "."+companyDomain) {
			return &Flag{
				Code:     codeEmailDomainMismatch,
				Severity: SeverityHard,
				Message:  "contact email domain does not match the company domain",
			}
		}
	}
	return nil
}

func companyDomain(j job.Job) string {
	if j.URL != "" {
		u, err := url.Parse(j.URL)
		if err == nil {
			host := strings.ToLower(strings.TrimPrefix(u.Hostname(), "www."))
			if host != "" && !strings.Contains(host, "greenhouse.io") && !strings.Contains(host, "lever.co") {
				return host
			}
		}
	}
	slug := slugCompany(j.Company)
	if slug == "" {
		return ""
	}
	return slug + ".com"
}

func slugCompany(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return ""
	}
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func checkPayToWork(text string) *Flag {
	if rePayToWork.MatchString(text) {
		return &Flag{
			Code:     codePayToWork,
			Severity: SeverityHard,
			Message:  "job description asks for payment, deposit, or fees",
		}
	}
	return nil
}

func checkIDBeforeInterview(text string) *Flag {
	if reIDBeforeInt.MatchString(text) || reIDBeforeInt2.MatchString(text) {
		return &Flag{
			Code:     codeIDBeforeInterview,
			Severity: SeverityHard,
			Message:  "asks for identity or banking details before an interview",
		}
	}
	return nil
}

func checkCloneAndRun(text string) *Flag {
	if reCloneRun.MatchString(text) {
		return &Flag{
			Code:     codeCloneAndRun,
			Severity: SeverityHard,
			Message:  "asks you to clone a repository and run untrusted code",
		}
	}
	if strings.Contains(text, "clone") && strings.Contains(text, "repo") &&
		(strings.Contains(text, "run locally") || strings.Contains(text, "execute") || strings.Contains(text, "npm install")) {
		return &Flag{
			Code:     codeCloneAndRun,
			Severity: SeverityHard,
			Message:  "asks you to clone a repository and run untrusted code",
		}
	}
	return nil
}

func checkContactWithoutDomain(j job.Job, text string) *Flag {
	hasMessenger := reWhatsApp.MatchString(text) || reTelegram.MatchString(text)
	if !hasMessenger {
		return nil
	}
	companyDomain := companyDomain(j)
	hasCompanySite := false
	if companyDomain != "" && strings.Contains(text, companyDomain) {
		hasCompanySite = true
	}
	for _, m := range reURLHost.FindAllStringSubmatch(text, -1) {
		host := strings.ToLower(m[1])
		if host != "" && !freeMailDomains[host] && !strings.Contains(host, "t.me") && !strings.Contains(host, "wa.me") {
			if companyDomain == "" || strings.Contains(host, strings.TrimSuffix(companyDomain, ".com")) {
				hasCompanySite = true
				break
			}
		}
	}
	if j.URL != "" {
		u, _ := url.Parse(j.URL)
		if u != nil && u.Hostname() != "" {
			hasCompanySite = true
		}
	}
	if !hasCompanySite {
		return &Flag{
			Code:     codeContactWithoutDomain,
			Severity: SeverityHard,
			Message:  "only messenger/personal contact with no company web presence",
		}
	}
	return nil
}

func checkVagueLocation(j job.Job, text string) *Flag {
	if j.Country != "" || (j.Region != "" && j.Region != "global") {
		return nil
	}
	vague := strings.Contains(text, "remote worldwide") || strings.Contains(text, "work from anywhere") ||
		strings.Contains(text, "global remote")
	hasOffice := strings.Contains(text, "office") || strings.Contains(text, "headquarters") ||
		strings.Contains(text, "based in") || strings.Contains(text, "location:")
	if vague && !hasOffice {
		return &Flag{
			Code:     codeVagueLocation,
			Severity: SeveritySoft,
			Message:  "no office or location, only broad remote wording",
		}
	}
	return nil
}

func checkDuplicateJD(j job.Job, ctx Context) *Flag {
	if ctx.PeerDescriptionHashes == nil || j.Description == "" {
		return nil
	}
	h := descriptionHash(j.Description)
	if ctx.PeerDescriptionHashes[h] {
		return &Flag{
			Code:     codeDuplicateJD,
			Severity: SeveritySoft,
			Message:  "description closely matches another posting",
		}
	}
	return nil
}

// DescriptionHash normalizes the first 400 characters for duplicate detection.
func DescriptionHash(desc string) string {
	return descriptionHash(desc)
}

func descriptionHash(desc string) string {
	norm := normalizeDesc(desc)
	sum := sha256.Sum256([]byte(norm))
	return hex.EncodeToString(sum[:8])
}

func normalizeDesc(desc string) string {
	desc = strings.ToLower(desc)
	if len(desc) > 400 {
		desc = desc[:400]
	}
	var b strings.Builder
	lastSpace := false
	for _, r := range desc {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastSpace = false
			continue
		}
		if !lastSpace {
			b.WriteByte(' ')
			lastSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

func checkSalaryOutlier(j job.Job) *Flag {
	if !j.SalaryDisclosed || j.SalaryMin <= 0 || j.SalaryCurrency != "USD" {
		return nil
	}
	band := salaryBandUSD(j.Seniority, j.Title)
	if band.max == 0 {
		return nil
	}
	if j.SalaryMin > band.max {
		return &Flag{
			Code:     codeSalaryOutlier,
			Severity: SeveritySoft,
			Message:  "stated compensation is far above typical range for the level",
		}
	}
	return nil
}

type salaryBand struct {
	max int
}

func salaryBandUSD(seniority, title string) salaryBand {
	s := strings.ToLower(seniority + " " + title)
	switch {
	case strings.Contains(s, "intern"):
		return salaryBand{max: 80_000}
	case strings.Contains(s, "junior"):
		return salaryBand{max: 120_000}
	case strings.Contains(s, "mid") || strings.Contains(s, "middle"):
		return salaryBand{max: 160_000}
	case strings.Contains(s, "senior"):
		return salaryBand{max: 220_000}
	case strings.Contains(s, "staff") || strings.Contains(s, "principal") || strings.Contains(s, "lead"):
		return salaryBand{max: 300_000}
	default:
		return salaryBand{max: 180_000}
	}
}

func checkUnnamedCompany(j job.Job) *Flag {
	name := strings.TrimSpace(strings.ToLower(j.Company))
	if name == "" {
		return &Flag{
			Code:     codeUnnamedCompany,
			Severity: SeveritySoft,
			Message:  "company name is missing",
		}
	}
	placeholders := []string{"confidential", "stealth", "undisclosed", "tbd", "n/a", "your company", "company name"}
	for _, p := range placeholders {
		if name == p || strings.Contains(name, p) {
			return &Flag{
				Code:     codeUnnamedCompany,
				Severity: SeveritySoft,
				Message:  "company name is missing or a placeholder",
			}
		}
	}
	return nil
}

// PeerHashesFromJobs builds a set of description hashes for all jobs except skipID.
func PeerHashesFromJobs(jobs []job.Job, skipID string) map[string]bool {
	out := map[string]bool{}
	for _, j := range jobs {
		if j.ID == skipID || j.Description == "" {
			continue
		}
		out[descriptionHash(j.Description)] = true
	}
	return out
}
