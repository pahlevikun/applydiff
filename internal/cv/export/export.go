package export

import (
	"fmt"
	"strings"

	"github.com/pahlevikun/applydiff/internal/cv"
)

// ToTeX renders a full LaTeX document from structured CV (ATS-oriented, matches importer patterns).
func ToTeX(doc *cv.Document) string {
	if doc == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("\\documentclass[11pt]{article}\n")
	b.WriteString("\\usepackage[margin=0.75in]{geometry}\n")
	b.WriteString("\\usepackage{enumitem}\n")
	b.WriteString("\\setlist[itemize]{noitemsep, topsep=2pt}\n")
	b.WriteString("\\pagestyle{empty}\n")
	b.WriteString("\\begin{document}\n\n")

	b.WriteString(fmt.Sprintf("{\\LARGE %s}\\\\\n", escape(doc.Contact.Name)))
	b.WriteString(fmt.Sprintf("{\\large %s}\\\\\n", escape(doc.Contact.Headline)))
	contact := strings.Join([]string{
		escape(doc.Contact.Email),
		escape(doc.Contact.Site),
		escape(doc.Contact.LinkedIn),
	}, " $\\cdot$ ")
	if contact != "" {
		b.WriteString(contact + "\\\\\n")
	}
	locLine := strings.TrimSpace(doc.Contact.Location)
	if doc.Contact.TimezoneLine != "" {
		if locLine != "" {
			locLine += " "
		}
		locLine += doc.Contact.TimezoneLine
	}
	if locLine != "" {
		b.WriteString(escape(locLine) + "\\\\\n")
	}
	b.WriteString("\n")

	b.WriteString("\\section*{Summary}\n")
	b.WriteString(escape(doc.Summary.Text) + "\n\n")

	b.WriteString("\\section*{Technical Skills}\n")
	for _, sg := range doc.Skills {
		b.WriteString(fmt.Sprintf("\\textbf{%s}: %s\\\\\n", escape(sg.Label), escape(strings.Join(sg.Items, ", "))))
	}
	b.WriteString("\n")

	b.WriteString("\\section*{Work Experience}\n")
	for _, exp := range doc.Experience {
		companyLine := exp.Company
		if exp.Team != "" {
			companyLine += " — " + exp.Team
		}
		b.WriteString(fmt.Sprintf("\\textbf{%s} \\hfill %s\\\\\n", escape(companyLine), escape(exp.Location)))
		end := "Present"
		if exp.End != nil && *exp.End != "" {
			end = *exp.End
		}
		b.WriteString(fmt.Sprintf("\\textit{%s} \\hfill %s -- %s\\\\\n", escape(exp.Title), escape(exp.Start), escape(end)))
		if len(exp.Bullets) > 0 {
			b.WriteString("\\begin{itemize}\n")
			for _, bullet := range exp.Bullets {
				b.WriteString("\\item " + escapePercentInText(bullet.Text) + "\n")
			}
			b.WriteString("\\end{itemize}\n")
		}
		b.WriteString("\n")
	}

	if len(doc.Education) > 0 {
		b.WriteString("\\section*{Education}\n")
		for _, ed := range doc.Education {
			line := fmt.Sprintf("%s, %s (%s)", ed.Degree, ed.School, ed.Year)
			b.WriteString(fmt.Sprintf("\\textbf{%s}\\\\\n", escape(line)))
		}
		b.WriteString("\n")
	}

	if len(doc.Achievements) > 0 {
		b.WriteString("\\section*{Achievements}\n")
		b.WriteString("\\begin{itemize}\n")
		for _, ach := range doc.Achievements {
			b.WriteString("\\item " + escapePercentInText(ach.Text) + "\n")
		}
		b.WriteString("\\end{itemize}\n")
	}

	b.WriteString("\\end{document}\n")
	return b.String()
}

func escape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\textbackslash{}")
	s = strings.ReplaceAll(s, "&", "\\&")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "#", "\\#")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

func escapePercentInText(s string) string {
	// Re-escape only % for bullet body (already plain text from JSON).
	return strings.ReplaceAll(escape(s), "\\textbackslash{}\\%", "\\%")
}

// RequiredSections returns section titles that must appear in exported TeX.
func RequiredSections() []string {
	return []string{"Summary", "Technical Skills", "Work Experience"}
}

// ContainsSections reports whether all required section headers exist.
func ContainsSections(tex string) bool {
	for _, sec := range RequiredSections() {
		if !strings.Contains(tex, "\\section*{"+sec+"}") {
			return false
		}
	}
	return true
}
