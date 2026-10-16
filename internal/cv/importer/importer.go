package importer

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pahlevikun/applydiff/internal/cv"
)

var (
	sectionRe     = regexp.MustCompile(`\\section\*\{([^}]+)\}`)
	skillLineRe   = regexp.MustCompile(`\\textbf\{([^}]+)\}:\s*(.+)`)
	factCommentRe = regexp.MustCompile(`%\s*fact:([a-zA-Z0-9_.-]+)=([^\s]+)`)
	percentRe     = regexp.MustCompile(`(\d+(?:\.\d+)?)\\%`)
)

// ImportTex parses a best-effort ATS LaTeX resume into cv.json + facts.json shapes.
func ImportTex(path string) (*cv.Document, *cv.Facts, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	return ImportTexBytes(data)
}

func ImportTexBytes(data []byte) (*cv.Document, *cv.Facts, error) {
	lines := strings.Split(string(data), "\n")
	doc := &cv.Document{
		Person: "person",
		Skills: []cv.SkillGroup{},
	}
	facts := &cv.Facts{Facts: map[string]string{}}

	var section string
	var currentExp *cv.Experience
	bulletIdx := 0
	expIdx := 0
	eduIdx := 0
	achIdx := 0

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if m := factCommentRe.FindStringSubmatch(line); len(m) == 3 {
			facts.Facts[m[1]] = strings.ReplaceAll(m[2], "\\%", "%")
		}
		if m := sectionRe.FindStringSubmatch(line); len(m) == 2 {
			section = strings.ToLower(strings.TrimSpace(m[1]))
			currentExp = nil
			continue
		}
		if strings.HasPrefix(line, "\\item") {
			text := cleanLatex(strings.TrimPrefix(line, "\\item"))
			text = strings.TrimSpace(text)
			if section == "work experience" && currentExp != nil {
				bulletIdx++
				id := fmt.Sprintf("exp.%d.b%d", expIdx, bulletIdx)
				currentExp.Bullets = append(currentExp.Bullets, cv.Bullet{ID: id, Text: text})
				extractFactsFromText(text, facts)
			}
			if section == "achievements" {
				achIdx++
				doc.Achievements = append(doc.Achievements, cv.Achievement{
					ID:   fmt.Sprintf("ach.%d", achIdx),
					Text: text,
				})
				extractFactsFromText(text, facts)
			}
			continue
		}
		if strings.HasPrefix(line, "\\textbf{") && section == "work experience" {
			companyTeam := extractBrace(line, "\\textbf{")
			loc := ""
			if i := strings.Index(line, "\\hfill"); i >= 0 {
				loc = cleanLatex(line[i+len("\\hfill"):])
			}
			parts := strings.SplitN(companyTeam, "—", 2)
			company := strings.TrimSpace(parts[0])
			team := ""
			if len(parts) == 2 {
				team = strings.TrimSpace(parts[1])
			}
			expIdx++
			bulletIdx = 0
			doc.Experience = append(doc.Experience, cv.Experience{
				ID:       fmt.Sprintf("exp.%d", expIdx),
				Company:  company,
				Team:     team,
				Location: loc,
				Bullets:  []cv.Bullet{},
			})
			currentExp = &doc.Experience[len(doc.Experience)-1]
			continue
		}
		if strings.Contains(line, "\\textit{") && section == "work experience" && currentExp != nil {
			title := extractBrace(line, "\\textit{")
			dates := ""
			if i := strings.Index(line, "\\hfill"); i >= 0 {
				dates = cleanLatex(line[i+len("\\hfill"):])
			}
			start, end := parseDateRange(dates)
			currentExp.Title = title
			currentExp.Start = start
			currentExp.End = end
			continue
		}
		if m := skillLineRe.FindStringSubmatch(line); len(m) == 3 && strings.Contains(section, "skill") {
			label := cleanLatex(m[1])
			items := splitSkillItems(cleanLatex(m[2]))
			doc.Skills = append(doc.Skills, cv.SkillGroup{
				ID:    fmt.Sprintf("skills.%s", slug(label)),
				Label: label,
				Items: items,
			})
			continue
		}
		if section == "summary" && line != "" && !strings.HasPrefix(line, "%") && !strings.HasPrefix(line, "\\") {
			doc.Summary = cv.Summary{ID: "summary", Text: cleanLatex(line)}
			continue
		}
		if section == "education" && strings.HasPrefix(line, "\\textbf{") {
			eduIdx++
			rest := cleanLatex(extractBrace(line, "\\textbf{"))
			degree, school, year := parseEducation(rest)
			doc.Education = append(doc.Education, cv.Education{
				ID:     fmt.Sprintf("edu.%d", eduIdx),
				School: school,
				Degree: degree,
				Year:   year,
			})
			continue
		}
		if strings.HasPrefix(line, "{\\LARGE") || strings.HasPrefix(line, "{\\large") {
			val := cleanLatex(extractBrace(line, "{"))
			if doc.Contact.Name == "" {
				doc.Contact.Name = val
				doc.Person = slugFromName(val)
			} else if doc.Contact.Headline == "" {
				doc.Contact.Headline = val
			}
			continue
		}
		if doc.Contact.Email == "" && strings.Contains(line, "@") {
			parseContactLine(line, doc)
		}
	}

	if doc.Summary.ID == "" {
		doc.Summary = cv.Summary{ID: "summary", Text: ""}
	}
	if err := cv.Validate(doc); err != nil {
		return doc, facts, err
	}
	return doc, facts, nil
}

func parseContactLine(line string, doc *cv.Document) {
	clean := cleanLatex(line)
	parts := strings.Split(clean, "·")
	for i, p := range parts {
		p = strings.TrimSpace(p)
		switch i {
		case 0:
			if strings.Contains(p, "@") {
				doc.Contact.Email = p
			}
		case 1:
			doc.Contact.Site = p
		case 2:
			doc.Contact.LinkedIn = p
		}
	}
	if strings.Contains(clean, "GMT") {
		doc.Contact.TimezoneLine = clean
	}
	if strings.Contains(clean, "Indonesia") && doc.Contact.Location == "" {
		loc := strings.Split(clean, "(")[0]
		doc.Contact.Location = strings.TrimSpace(loc)
	}
}

func parseEducation(s string) (degree, school, year string) {
	if i := strings.LastIndex(s, "("); i >= 0 && strings.HasSuffix(s, ")") {
		year = strings.TrimSuffix(s[i+1:], ")")
		s = strings.TrimSpace(s[:i])
	}
	if j := strings.Index(s, ","); j >= 0 {
		degree = strings.TrimSpace(s[:j])
		school = strings.TrimSpace(s[j+1:])
	} else {
		school = s
	}
	return degree, school, year
}

func parseDateRange(s string) (start string, end *string) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "—", "--")
	s = strings.ReplaceAll(s, "–", "--")
	parts := strings.Split(s, "--")
	if len(parts) >= 1 {
		start = normalizeMonth(parts[0])
	}
	if len(parts) >= 2 {
		p2 := strings.TrimSpace(parts[1])
		if strings.EqualFold(p2, "present") {
			return start, nil
		}
		endStr := normalizeMonth(p2)
		return start, &endStr
	}
	return start, nil
}

func normalizeMonth(s string) string {
	s = strings.TrimSpace(s)
	months := map[string]string{
		"jan": "01", "feb": "02", "mar": "03", "apr": "04", "may": "05", "jun": "06",
		"jul": "07", "aug": "08", "sep": "09", "oct": "10", "nov": "11", "dec": "12",
	}
	lower := strings.ToLower(s)
	for name, num := range months {
		if strings.HasPrefix(lower, name) {
			rest := strings.TrimSpace(s[3:])
			if len(rest) >= 4 {
				return rest + "-" + num
			}
		}
	}
	return s
}

func extractFactsFromText(text string, facts *cv.Facts) {
	for _, m := range percentRe.FindAllStringSubmatch(text, -1) {
		val := m[1] + "%"
		key := "metric_" + strings.ReplaceAll(m[1], ".", "_")
		if facts.Facts[key] == "" {
			facts.Facts[key] = val
		}
	}
}

func splitSkillItems(s string) []string {
	s = strings.ReplaceAll(s, ";", ",")
	var items []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			items = append(items, p)
		}
	}
	return items
}

func extractBrace(line, prefix string) string {
	i := strings.Index(line, prefix)
	if i < 0 {
		return ""
	}
	rest := line[i+len(prefix):]
	end := strings.Index(rest, "}")
	if end < 0 {
		return cleanLatex(rest)
	}
	return cleanLatex(rest[:end])
}

func cleanLatex(s string) string {
	s = strings.ReplaceAll(s, "\\&", "&")
	s = strings.ReplaceAll(s, "\\%", "%")
	s = strings.ReplaceAll(s, "~", " ")
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")
	s = strings.ReplaceAll(s, "\\hfill", "")
	s = strings.ReplaceAll(s, "\\textit", "")
	s = strings.ReplaceAll(s, "\\textbf", "")
	s = strings.ReplaceAll(s, "\\LARGE", "")
	s = strings.ReplaceAll(s, "\\large", "")
	return strings.TrimSpace(s)
}

func slugFromName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "person"
	}
	return out
}

func slug(s string) string {
	return slugFromName(s)
}

// WriteJSON writes cv and facts to paths (helper for cmd/import).
func WriteJSON(cvPath, factsPath string, doc *cv.Document, facts *cv.Facts) error {
	cvData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	factsWrap := map[string]map[string]string{"facts": facts.Facts}
	factsData, err := json.MarshalIndent(factsWrap, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(cvPath, cvData, 0o644); err != nil {
		return err
	}
	return os.WriteFile(factsPath, factsData, 0o644)
}
