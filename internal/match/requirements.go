package match

import (
	"regexp"
	"strings"

	"github.com/pahlevikun/applydiff/internal/cv"
)

var requirementLineRe = regexp.MustCompile(`(?i)^[\s*\-•]+(.+)$`)

// ExtractRequirements pulls requirement phrases from JD text and classifies each against the CV.
func ExtractRequirements(jd string, doc *cv.Document) []Requirement {
	phrases := extractPhrases(jd)
	out := make([]Requirement, 0, len(phrases))
	for _, phrase := range phrases {
		out = append(out, classifyRequirement(phrase, doc))
	}
	return out
}

func extractPhrases(jd string) []string {
	var phrases []string
	seen := make(map[string]bool)
	for _, line := range strings.Split(jd, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		candidate := line
		if m := requirementLineRe.FindStringSubmatch(line); len(m) == 2 {
			candidate = strings.TrimSpace(m[1])
		}
		lower := strings.ToLower(candidate)
		if !looksLikeRequirement(lower) && !strings.Contains(lower, "years") {
			continue
		}
		key := strings.ToLower(candidate)
		if seen[key] {
			continue
		}
		seen[key] = true
		phrases = append(phrases, candidate)
	}
	if len(phrases) == 0 {
		for _, tok := range requirementTokens(jd) {
			phrases = append(phrases, tok)
		}
	}
	return phrases
}

func looksLikeRequirement(lower string) bool {
	markers := []string{
		"required", "must have", "need", "experience with", "proficient",
		"strong", "familiar", "knowledge of", "skills:", "requirements:",
	}
	for _, m := range markers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

func classifyRequirement(phrase string, doc *cv.Document) Requirement {
	tokens := tokenizePhrase(phrase)
	if len(tokens) == 0 {
		return Requirement{Phrase: phrase, Status: StatusGap}
	}
	if doc == nil {
		return Requirement{Phrase: phrase, Status: StatusGap}
	}

	bestStatus := StatusGap
	var bestID, bestText string

	for _, tok := range tokens {
		id, text, status := matchToken(tok, doc)
		if statusRank(status) > statusRank(bestStatus) {
			bestStatus = status
			bestID = id
			bestText = text
		}
	}
	if bestStatus == StatusGap {
		return Requirement{Phrase: phrase, Status: StatusGap}
	}
	return Requirement{
		Phrase:      phrase,
		Status:      bestStatus,
		CVFieldID:   bestID,
		CVFieldText: bestText,
	}
}

func statusRank(s RequirementStatus) int {
	switch s {
	case StatusMatch:
		return 3
	case StatusPartial:
		return 2
	default:
		return 1
	}
}

func matchToken(tok string, doc *cv.Document) (fieldID, fieldText string, status RequirementStatus) {
	norm := normalizeToken(tok)
	if norm == "" {
		return "", "", StatusGap
	}
	for _, sg := range doc.Skills {
		for _, item := range sg.Items {
			itemNorm := normalizeToken(item)
			if itemNorm == norm {
				val, _ := doc.Field(sg.ID)
				return sg.ID, val, StatusMatch
			}
			if synonymOverlap(itemNorm, norm) {
				val, _ := doc.Field(sg.ID)
				return sg.ID, val, StatusPartial
			}
		}
	}
	corpus := cvCorpus(doc)
	if strings.Contains(strings.ToLower(corpus), norm) {
		return doc.Summary.ID, doc.Summary.Text, StatusPartial
	}
	return "", "", StatusGap
}

func tokenizePhrase(phrase string) []string {
	phrase = strings.ToLower(phrase)
	for _, prefix := range []string{
		"required:", "required ", "must have ", "experience with ", "proficient in ",
		"strong ", "knowledge of ",
	} {
		if strings.HasPrefix(phrase, prefix) {
			phrase = strings.TrimPrefix(phrase, prefix)
			break
		}
	}
	if idx := strings.LastIndex(phrase, " with "); idx >= 0 {
		phrase = phrase[idx+len(" with "):]
	}
	var tokens []string
	for _, part := range strings.FieldsFunc(phrase, func(r rune) bool {
		return r == ',' || r == ';' || r == '/' || r == '(' || r == ')' || r == '+'
	}) {
		t := normalizeToken(part)
		if len(t) >= 2 && !stopWords[t] {
			tokens = append(tokens, t)
		}
	}
	return tokens
}

func requirementTokens(jd string) []string {
	words := significantWords(jd)
	var tokens []string
	for w := range words {
		if len(w) >= 2 && !stopWords[w] {
			tokens = append(tokens, w)
		}
	}
	for _, short := range shortTechTokens {
		if strings.Contains(strings.ToLower(jd), short) {
			tokens = append(tokens, short)
		}
	}
	return tokens
}

var shortTechTokens = []string{"go", "js", "c#", "c++", "ai", "ml", "qa", "ux", "ui", "k8s"}
var shortTechSet = func() map[string]bool {
	m := make(map[string]bool)
	for _, t := range shortTechTokens {
		m[t] = true
	}
	return m
}()

var stopWords = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "you": true,
	"our": true, "will": true, "have": true, "are": true, "this": true,
	"that": true, "from": true, "your": true, "we": true, "a": true, "an": true,
}

func significantWords(text string) map[string]bool {
	out := make(map[string]bool)
	for _, w := range strings.Fields(strings.ToLower(text)) {
		w = strings.Trim(w, ".,;:!?\"'")
		if len(w) < 2 || stopWords[w] {
			continue
		}
		if len(w) == 2 && !shortTechSet[w] {
			continue
		}
		out[w] = true
	}
	return out
}

func normalizeToken(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Trim(s, ".,;:!?\"'")
	return s
}

var synonymTable = map[string][]string{
	"go":        {"golang"},
	"golang":    {"go"},
	"javascript": {"js"},
	"js":        {"javascript"},
	"k8s":       {"kubernetes"},
	"kubernetes": {"k8s"},
	"postgres":  {"postgresql"},
	"postgresql": {"postgres"},
}

func synonyms(tok string) []string {
	return synonymTable[tok]
}

func synonymOverlap(a, b string) bool {
	if a == b {
		return true
	}
	for _, s := range synonyms(a) {
		if s == b {
			return true
		}
	}
	for _, s := range synonyms(b) {
		if s == a {
			return true
		}
	}
	return false
}
