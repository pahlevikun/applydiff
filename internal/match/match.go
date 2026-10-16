package match

import (
	"strings"
	"time"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/job"
)

// Dimension scores are each 0–5; Total is 0–25.
type Dimensions struct {
	Skills     int `json:"skills"`
	Seniority  int `json:"seniority"`
	Domain     int `json:"domain"`
	Logistics  int `json:"logistics"`
	Motivation int `json:"motivation"`
}

func (d Dimensions) Total() int {
	return d.Skills + d.Seniority + d.Domain + d.Logistics + d.Motivation
}

type RequirementStatus string

const (
	StatusMatch   RequirementStatus = "match"
	StatusPartial RequirementStatus = "partial"
	StatusGap     RequirementStatus = "gap"
)

type Requirement struct {
	Phrase      string            `json:"phrase"`
	Status      RequirementStatus `json:"status"`
	CVFieldID   string            `json:"cv_field_id,omitempty"`
	CVFieldText string            `json:"cv_field_text,omitempty"`
}

type Legitimacy struct {
	CompanyVerified  bool     `json:"company_verified"`
	OfficialATSURL   bool     `json:"official_ats_url"`
	SalaryConsistent bool     `json:"salary_consistent"`
	Flags            []string `json:"flags,omitempty"`
}

type Input struct {
	Job job.Job
	CV  *cv.Document
}

type Result struct {
	JobID        string        `json:"job_id"`
	Total        int           `json:"total"`
	Dimensions   Dimensions    `json:"dimensions"`
	Requirements []Requirement `json:"requirements"`
	Legitimacy   Legitimacy    `json:"legitimacy"`
	JDText       string        `json:"jd_text"`
	ComputedAt   string        `json:"computed_at"`
}

// Evaluate scores the job against the CV using deterministic heuristics.
func Evaluate(in Input) Result {
	jd := jdText(in.Job)
	reqs := ExtractRequirements(jd, in.CV)
	leg := CheckLegitimacy(in.Job)
	dims := scoreDimensions(in.Job, jd, in.CV)
	now := time.Now().UTC().Format(time.RFC3339)
	return Result{
		JobID:        in.Job.ID,
		Total:        dims.Total(),
		Dimensions:   dims,
		Requirements: reqs,
		Legitimacy:   leg,
		JDText:       jd,
		ComputedAt:   now,
	}
}

func jdText(j job.Job) string {
	if strings.TrimSpace(j.Description) != "" {
		return strings.TrimSpace(j.Description)
	}
	var parts []string
	if j.Title != "" {
		parts = append(parts, j.Title)
	}
	if j.Company != "" {
		parts = append(parts, "at "+j.Company)
	}
	return strings.Join(parts, " ")
}

func scoreDimensions(j job.Job, jd string, doc *cv.Document) Dimensions {
	return Dimensions{
		Skills:     scoreSkills(jd, doc),
		Seniority:  scoreSeniority(j.Title, doc),
		Domain:     scoreDomain(jd, doc),
		Logistics:  scoreLogistics(j, doc),
		Motivation: scoreMotivation(jd, doc),
	}
}

func scoreSkills(jd string, doc *cv.Document) int {
	if doc == nil {
		return 0
	}
	jdLower := strings.ToLower(jd)
	var skillNames []string
	for _, sg := range doc.Skills {
		for _, item := range sg.Items {
			n := normalizeToken(item)
			if n != "" {
				skillNames = append(skillNames, n)
			}
		}
	}
	if len(skillNames) == 0 {
		return 0
	}
	matched := 0
	for _, sk := range skillNames {
		if strings.Contains(jdLower, sk) {
			matched++
			continue
		}
		for _, syn := range synonyms(sk) {
			if strings.Contains(jdLower, syn) {
				matched++
				break
			}
		}
	}
	ratio := float64(matched) / float64(len(skillNames))
	return clamp5(int(ratio*5 + 0.49))
}

func scoreSeniority(jobTitle string, doc *cv.Document) int {
	if doc == nil {
		return 0
	}
	jobLevel := seniorityRank(jobTitle)
	cvLevel := seniorityRank(doc.Contact.Headline)
	for _, exp := range doc.Experience {
		if exp.End == nil {
			cvLevel = max(cvLevel, seniorityRank(exp.Title))
		}
	}
	diff := jobLevel - cvLevel
	switch {
	case diff <= 0:
		return 5
	case diff == 1:
		return 3
	case diff == 2:
		return 2
	default:
		return 1
	}
}

func seniorityRank(title string) int {
	t := strings.ToLower(title)
	switch {
	case strings.Contains(t, "principal"), strings.Contains(t, "staff"):
		return 5
	case strings.Contains(t, "lead"), strings.Contains(t, "sr."), strings.Contains(t, "senior"):
		return 4
	case strings.Contains(t, "mid"), strings.Contains(t, "ii"):
		return 3
	case strings.Contains(t, "junior"), strings.Contains(t, "jr."), strings.Contains(t, "entry"):
		return 1
	default:
		return 3
	}
}

func scoreDomain(jd string, doc *cv.Document) int {
	if doc == nil {
		return 0
	}
	jdWords := significantWords(jd)
	if len(jdWords) == 0 {
		return 3
	}
	cvText := strings.ToLower(cvCorpus(doc))
	hits := 0
	for w := range jdWords {
		if strings.Contains(cvText, w) {
			hits++
		}
	}
	ratio := float64(hits) / float64(len(jdWords))
	return clamp5(int(ratio*5 + 0.49))
}

func scoreLogistics(j job.Job, doc *cv.Document) int {
	score := 3
	if j.Remote == "yes" {
		score += 1
	}
	if doc != nil && strings.TrimSpace(doc.Contact.TimezoneLine) != "" {
		tz := strings.ToLower(doc.Contact.TimezoneLine)
		jd := strings.ToLower(jdText(j))
		if strings.Contains(jd, "eu") || strings.Contains(jd, "europe") {
			if strings.Contains(tz, "eu") || strings.Contains(tz, "overlap") {
				score += 1
			}
		} else {
			score += 1
		}
	}
	if strings.Contains(strings.ToLower(jdText(j)), "on-site only") && j.Remote != "yes" {
		score -= 2
	}
	return clamp5(score)
}

func scoreMotivation(jd string, doc *cv.Document) int {
	if doc == nil {
		return 0
	}
	needles := []string{"mission", "impact", "product", "customer", "scale", "growth", "open source", "fintech", "health"}
	jdLower := strings.ToLower(jd)
	cvLower := strings.ToLower(cvCorpus(doc))
	hits := 0
	for _, n := range needles {
		if strings.Contains(jdLower, n) && strings.Contains(cvLower, n) {
			hits++
		}
	}
	if hits >= 3 {
		return 5
	}
	if hits == 2 {
		return 4
	}
	if hits == 1 {
		return 3
	}
	return 2
}

func cvCorpus(doc *cv.Document) string {
	var b strings.Builder
	b.WriteString(doc.Summary.Text)
	b.WriteByte(' ')
	b.WriteString(doc.Contact.Headline)
	for _, exp := range doc.Experience {
		b.WriteByte(' ')
		b.WriteString(exp.Title)
		b.WriteByte(' ')
		b.WriteString(exp.Company)
		for _, bullet := range exp.Bullets {
			b.WriteByte(' ')
			b.WriteString(bullet.Text)
		}
	}
	for _, ach := range doc.Achievements {
		b.WriteByte(' ')
		b.WriteString(ach.Text)
	}
	return b.String()
}

func cvSkillSet(doc *cv.Document) map[string]bool {
	out := make(map[string]bool)
	for _, sg := range doc.Skills {
		for _, item := range sg.Items {
			n := normalizeToken(item)
			if n != "" {
				out[n] = true
			}
			for _, syn := range synonyms(n) {
				out[syn] = true
			}
		}
	}
	return out
}

func clamp5(n int) int {
	if n < 0 {
		return 0
	}
	if n > 5 {
		return 5
	}
	return n
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
