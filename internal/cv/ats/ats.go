package ats

import (
	"os/exec"
	"regexp"
	"strings"
)

var sectionOrder = []string{
	"summary",
	"technical skills",
	"work experience",
	"education",
	"achievements",
}

var wordRe = regexp.MustCompile(`[A-Za-z][A-Za-z0-9+.#-]*`)

// PlainTextFromTeX strips LaTeX markup for ATS-style reading-order checks.
func PlainTextFromTeX(tex string) string {
	s := tex
	s = regexp.MustCompile(`(?m)^%.*$`).ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\\%", "%")
	s = regexp.MustCompile(`\\[a-zA-Z]+\*?(\[[^\]]*\])?`).ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")
	s = strings.ReplaceAll(s, "\\\\", "\n")
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// ReadingOrder checks that major sections appear in canonical order in plain text.
func ReadingOrder(plain string) (bool, []string) {
	lower := strings.ToLower(plain)
	var positions []int
	for _, sec := range sectionOrder {
		idx := strings.Index(lower, sec)
		if idx < 0 {
			continue
		}
		positions = append(positions, idx)
	}
	for i := 1; i < len(positions); i++ {
		if positions[i] < positions[i-1] {
			return false, sectionOrder
		}
	}
	return true, sectionOrder
}

// KeywordCoverage reports how many requested keywords appear in plain text (case-insensitive).
func KeywordCoverage(plain string, keywords []string) (matched []string, missing []string) {
	lower := strings.ToLower(plain)
	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(kw)) {
			matched = append(matched, kw)
		} else {
			missing = append(missing, kw)
		}
	}
	return matched, missing
}

// ExtractTokens returns lowercased word tokens useful for quick ATS checks.
func ExtractTokens(plain string) []string {
	return wordRe.FindAllString(strings.ToLower(plain), -1)
}

// PlainTextFromPDF uses pdftotext when available; otherwise returns empty string and false.
func PlainTextFromPDF(pdfPath string) (string, bool) {
	cmd := exec.Command("pdftotext", "-layout", pdfPath, "-")
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return string(out), true
}

// Report bundles ATS helper output.
type Report struct {
	Plain           string
	OrderOK         bool
	SectionOrder    []string
	KeywordsMatched []string
	KeywordsMissing []string
}

func AnalyzeTeX(tex string, keywords []string) Report {
	plain := PlainTextFromTeX(tex)
	ok, order := ReadingOrder(plain)
	matched, missing := KeywordCoverage(plain, keywords)
	return Report{
		Plain:           plain,
		OrderOK:         ok,
		SectionOrder:    order,
		KeywordsMatched: matched,
		KeywordsMissing: missing,
	}
}
