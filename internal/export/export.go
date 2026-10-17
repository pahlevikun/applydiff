package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pahlevikun/applydiff/internal/cv"
	cvexport "github.com/pahlevikun/applydiff/internal/cv/export"
	"github.com/pahlevikun/applydiff/internal/job"
)

// Result holds paths and compile metadata for one export run.
type Result struct {
	Dir           string `json:"dir"`
	CVTeX         string `json:"cv_tex"`
	CVPDF         string `json:"cv_pdf,omitempty"`
	CoverLetter   string `json:"cover_letter_tex"`
	ProofPage     string `json:"proof_page_tex"`
	PDFCompiled   bool   `json:"pdf_compiled"`
	CompileNote   string `json:"compile_note,omitempty"`
	CVSnapshot    string `json:"cv_snapshot"`
}

type Options struct {
	OutRoot string // default output/
	Job     job.Job
	Doc     *cv.Document
	Facts   *cv.Facts
}

// Build writes CV TeX drafts to a temp dir and the CV PDF under output/<date>-<company>-<role>.pdf.
func Build(opts Options) (*Result, error) {
	if opts.Doc == nil {
		return nil, fmt.Errorf("export: cv document required")
	}
	if opts.Job.ID == "" {
		return nil, fmt.Errorf("export: job id required")
	}
	root := opts.OutRoot
	if root == "" {
		root = "output"
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}

	workDir, err := os.MkdirTemp("", "applydiff-export-*")
	if err != nil {
		return nil, err
	}

	slug := slugify(opts.Job.Company)
	date := time.Now().UTC().Format("2006-01-02")

	cvTex := cvexport.ToTeX(opts.Doc)
	cvTexPath := filepath.Join(workDir, fmt.Sprintf("cv-%s-%s.tex", slug, date))
	if err := os.WriteFile(cvTexPath, []byte(cvTex), 0o644); err != nil {
		os.RemoveAll(workDir)
		return nil, err
	}

	cover := CoverLetterTeX(opts.Doc, opts.Job)
	if err := gateText(opts.Facts, coverPlainText(opts.Doc, opts.Job)); err != nil {
		os.RemoveAll(workDir)
		return nil, fmt.Errorf("export: cover letter: %w", err)
	}
	coverPath := filepath.Join(workDir, fmt.Sprintf("cover-%s-%s.tex", slug, date))
	if err := os.WriteFile(coverPath, []byte(cover), 0o644); err != nil {
		os.RemoveAll(workDir)
		return nil, err
	}

	proof := ProofPageTeX(opts.Doc)
	if err := gateText(opts.Facts, proofPlainText(opts.Doc)); err != nil {
		os.RemoveAll(workDir)
		return nil, fmt.Errorf("export: proof page: %w", err)
	}
	proofPath := filepath.Join(workDir, fmt.Sprintf("proof-%s-%s.tex", slug, date))
	if err := os.WriteFile(proofPath, []byte(proof), 0o644); err != nil {
		os.RemoveAll(workDir)
		return nil, err
	}

	snap, _ := json.Marshal(opts.Doc)
	res := &Result{
		Dir:         workDir,
		CVTeX:       cvTexPath,
		CoverLetter: coverPath,
		ProofPage:   proofPath,
		CVSnapshot:  string(snap),
	}

	pdfPath := PDFOutputPath(root, opts.Job, time.Now().UTC())
	if err := CompilePDF(cvTexPath, workDir, pdfPath); err != nil {
		res.CompileNote = err.Error()
	} else {
		res.CVPDF = pdfPath
		res.PDFCompiled = true
	}
	return res, nil
}

func gateText(facts *cv.Facts, text string) error {
	if facts == nil || len(facts.Facts) == 0 {
		return nil
	}
	return facts.GateNumberIn(text)
}

// CoverLetterTeX builds a short cover letter from CV contact + summary (no invented metrics).
func coverPlainText(doc *cv.Document, j job.Job) string {
	parts := []string{
		j.Company, j.Title, doc.Summary.Text, doc.Contact.Headline,
		doc.Contact.Name, doc.Contact.Email, doc.Contact.Site,
	}
	return strings.Join(parts, " ")
}

func proofPlainText(doc *cv.Document) string {
	var parts []string
	parts = append(parts, doc.Contact.Name, doc.Contact.Headline, doc.Contact.Location)
	parts = append(parts, topResults(doc, 3)...)
	parts = append(parts, doc.Contact.GitHub, doc.Contact.LinkedIn, doc.Contact.Site, doc.Contact.Email)
	return strings.Join(parts, " ")
}

func CoverLetterTeX(doc *cv.Document, j job.Job) string {
	var b strings.Builder
	b.WriteString("\\documentclass[11pt]{article}\n")
	b.WriteString("\\usepackage[margin=1in]{geometry}\n")
	b.WriteString("\\usepackage{parskip}\n")
	b.WriteString("\\begin{document}\n\n")
	b.WriteString(fmt.Sprintf("Dear %s hiring team,\\\\\n\n", escape(j.Company)))
	b.WriteString("I am writing to express interest in the ")
	b.WriteString(escape(j.Title))
	b.WriteString(" role. ")
	b.WriteString(escape(strings.TrimSpace(doc.Summary.Text)))
	b.WriteString("\n\n")
	b.WriteString("My background as ")
	b.WriteString(escape(doc.Contact.Headline))
	b.WriteString(" aligns with the requirements you outlined. I would welcome a conversation about how I can contribute.\n\n")
	b.WriteString("Best regards,\\\\\n")
	b.WriteString(escape(doc.Contact.Name) + "\\\\\n")
	if doc.Contact.Email != "" {
		b.WriteString(escape(doc.Contact.Email) + "\\\\\n")
	}
	if doc.Contact.Site != "" {
		b.WriteString(escape(doc.Contact.Site) + "\n")
	}
	b.WriteString("\n\\end{document}\n")
	return b.String()
}

// ProofPageTeX is a one-page portfolio from structured CV facts only.
func ProofPageTeX(doc *cv.Document) string {
	var b strings.Builder
	b.WriteString("\\documentclass[11pt]{article}\n")
	b.WriteString("\\usepackage[margin=0.75in]{geometry}\n")
	b.WriteString("\\usepackage{hyperref}\n")
	b.WriteString("\\pagestyle{empty}\n")
	b.WriteString("\\begin{document}\n\n")
	identity := strings.TrimSpace(doc.Contact.Headline)
	if doc.Contact.Location != "" {
		identity += " · " + doc.Contact.Location
	}
	b.WriteString(fmt.Sprintf("{\\LARGE %s}\\\\\n", escape(doc.Contact.Name)))
	b.WriteString(fmt.Sprintf("{\\large %s}\\\\\n\n", escape(identity)))

	b.WriteString("\\section*{Results}\n")
	b.WriteString("\\begin{itemize}\n")
	for _, line := range topResults(doc, 3) {
		b.WriteString("\\item " + escapePercentInText(line) + "\n")
	}
	b.WriteString("\\end{itemize}\n\n")

	b.WriteString("\\section*{Links}\n")
	links := []string{doc.Contact.GitHub, doc.Contact.LinkedIn, doc.Contact.Site}
	for _, l := range links {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("\\href{%s}{%s}\\\\\n", escapeURL(l), escape(l)))
	}
	b.WriteString("\n\\section*{Contact}\n")
	if doc.Contact.Email != "" {
		b.WriteString(escape(doc.Contact.Email) + "\\\\\n")
	}
	if doc.Contact.TimezoneLine != "" {
		b.WriteString(escape(doc.Contact.TimezoneLine) + "\n")
	}
	b.WriteString("\n\\end{document}\n")
	return b.String()
}

func topResults(doc *cv.Document, n int) []string {
	var out []string
	for _, a := range doc.Achievements {
		t := strings.TrimSpace(a.Text)
		if t != "" {
			out = append(out, t)
		}
	}
	for _, exp := range doc.Experience {
		for _, bullet := range exp.Bullets {
			t := strings.TrimSpace(bullet.Text)
			if t != "" {
				out = append(out, t)
			}
		}
	}
	if len(out) > n {
		out = out[:n]
	}
	return out
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
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
	return strings.ReplaceAll(escape(s), "\\textbackslash{}\\%", "\\%")
}

func escapeURL(s string) string {
	return strings.ReplaceAll(s, "%", "\\%")
}
