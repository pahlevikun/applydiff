package interview

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pahlevikun/applydiff/internal/job"
)

// PrepExportPaths are markdown and optional PDF paths under output/.
type PrepExportPaths struct {
	Markdown string
	PDF      string
}

// PrepOutputPath returns output/<YYYY-MM-DD>-<company>-<role>-interview.md
func PrepOutputPath(outRoot string, j job.Job, at time.Time) string {
	if outRoot == "" {
		outRoot = "output"
	}
	date := at.UTC().Format("2006-01-02")
	name := fmt.Sprintf("%s-%s-%s-interview.md", date, slugPart(j.Company), slugPart(j.Title))
	return filepath.Join(outRoot, name)
}

// PrepPDFPath replaces .md with .pdf for the same basename.
func PrepPDFPath(mdPath string) string {
	if strings.HasSuffix(mdPath, ".md") {
		return mdPath[:len(mdPath)-3] + ".pdf"
	}
	return mdPath + ".pdf"
}

func slugPart(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "role"
	}
	return s
}

// PrepDocument is input for markdown export.
type PrepDocument struct {
	Job         job.Job
	Loop        []string
	Topics      []string
	Questions   []string
	Stories     []Story
	Sessions    []PrepSessionExport
	WeakSteps   []WeakStep
	RubricLines []string
}

// PrepSessionExport is one mock session in the export.
type PrepSessionExport struct {
	Session MockSession
	Steps   []MockStep
	Scores  []RubricScore
}

// WeakStep highlights a step that timed out without output.
type WeakStep struct {
	SessionType string
	StepName    string
	Note        string
}

// RenderPrepMarkdown builds the interview prep sheet markdown.
func RenderPrepMarkdown(doc PrepDocument) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Interview prep — %s @ %s\n\n", doc.Job.Title, doc.Job.Company))
	b.WriteString(fmt.Sprintf("Job ID: `%s`\n\n", doc.Job.ID))

	b.WriteString("## Interview loop\n\n")
	if len(doc.Loop) == 0 {
		b.WriteString("_No loop generated — run `applydiff interview kit`._\n\n")
	} else {
		for i, stage := range doc.Loop {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, stage))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Topics\n\n")
	if len(doc.Topics) == 0 {
		b.WriteString("_None_\n\n")
	} else {
		for _, t := range doc.Topics {
			b.WriteString(fmt.Sprintf("- %s\n", t))
		}
		b.WriteString("\n")
	}

	if len(doc.Questions) > 0 {
		b.WriteString("## Questions to ask\n\n")
		for _, q := range doc.Questions {
			b.WriteString(fmt.Sprintf("- %s\n", q))
		}
		b.WriteString("\n")
	}

	for _, sess := range doc.Sessions {
		b.WriteString(fmt.Sprintf("## Mock session — %s\n\n", sess.Session.Type))
		if sess.Session.Prompt != "" {
			b.WriteString(sess.Session.Prompt + "\n\n")
		}
		for _, st := range sess.Steps {
			flag := ""
			if st.Weak {
				flag = " **weak**"
			}
			b.WriteString(fmt.Sprintf("### Step %d: %s%s\n\n", st.Idx+1, st.Name, flag))
			if st.Output != "" {
				b.WriteString("**Output:** " + st.Output + "\n\n")
			}
			if st.Note != "" {
				b.WriteString("**Notes:** " + st.Note + "\n\n")
			}
		}
		if len(sess.Scores) > 0 {
			b.WriteString("**Rubric**\n\n")
			for _, sc := range sess.Scores {
				line := fmt.Sprintf("- %s: %d/5", sc.Dimension, sc.Score)
				if sc.Comment != "" {
					line += " — " + sc.Comment
				}
				b.WriteString(line + "\n")
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("## Story bank\n\n")
	if len(doc.Stories) == 0 {
		b.WriteString("_No stories yet._\n\n")
	} else {
		byCat := map[string][]Story{}
		for _, st := range doc.Stories {
			byCat[st.Category] = append(byCat[st.Category], st)
		}
		for _, cat := range StoryCategories {
			items := byCat[cat]
			b.WriteString(fmt.Sprintf("### %s\n\n", cat))
			if len(items) == 0 {
				b.WriteString("_Empty category — add a STAR story._\n\n")
				continue
			}
			for _, st := range items {
				b.WriteString(fmt.Sprintf("- **S:** %s\n", st.Situation))
				b.WriteString(fmt.Sprintf("  **T:** %s\n", st.Task))
				b.WriteString(fmt.Sprintf("  **A:** %s\n", st.Action))
				b.WriteString(fmt.Sprintf("  **R:** %s\n", st.Result))
				if st.Learning != "" {
					b.WriteString(fmt.Sprintf("  **Learning:** %s\n", st.Learning))
				}
			}
			b.WriteString("\n")
		}
	}

	if len(doc.WeakSteps) > 0 {
		b.WriteString("## Weak spots\n\n")
		for _, w := range doc.WeakSteps {
			b.WriteString(fmt.Sprintf("- [%s] %s", w.SessionType, w.StepName))
			if w.Note != "" {
				b.WriteString(": " + w.Note)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

// WritePrepMarkdown writes markdown to path, creating parent dirs.
func WritePrepMarkdown(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
