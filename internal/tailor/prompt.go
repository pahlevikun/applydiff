package tailor

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pahlevikun/applydiff/internal/cv"
	"github.com/pahlevikun/applydiff/internal/job"
)

// BuildPrompt assembles the drafter prompt from CV, JD, and facts.
func BuildPrompt(j job.Job, doc *cv.Document, facts *cv.Facts) string {
	var b strings.Builder
	b.WriteString("You tailor a structured CV for a job posting. Output ONLY valid JSON:\n")
	b.WriteString(`{"hunks":[{"field_id":"...","section":"...","before":"...","after":"...","rationale":"...","evidence":[]}]}` + "\n\n")
	b.WriteString("Rules:\n")
	b.WriteString("- Rewrite, do not translate word-for-word. CV-English for English postings.\n")
	b.WriteString("- Each experience bullet starts with a past-tense action verb.\n")
	b.WriteString("- Mirror the posting's wording when the concept is already true; never add skills the person lacks.\n")
	b.WriteString("- Never invent numbers or metrics. Only use numbers that appear in approved facts or the source field text.\n")
	b.WriteString("- Do not use em dashes or words: leverage, utilize, robust, synergy.\n")
	b.WriteString(`- Fence (verbatim): "Jangan nambahin pengalaman, angka, atau skill yang nggak ada di teks aku."\n`)
	b.WriteString("- Never translate proper nouns (company, school, official titles).\n")
	b.WriteString("- Each hunk targets one field_id; before must match the current field value.\n\n")

	b.WriteString("## Job\n")
	b.WriteString(fmt.Sprintf("Company: %s\nTitle: %s\n\n", j.Company, j.Title))
	b.WriteString(j.Description)
	b.WriteString("\n\n## Approved facts (metrics only from here)\n")
	if facts != nil {
		for k, v := range facts.Facts {
			b.WriteString(fmt.Sprintf("%s: %s\n", k, v))
		}
	}
	b.WriteString("\n## Structured CV (JSON)\n")
	data, _ := json.Marshal(doc)
	b.Write(data)
	return b.String()
}
