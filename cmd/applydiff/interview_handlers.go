package main

import (
	"fmt"
	"strings"

	"github.com/pahlevikun/applydiff/internal/interview"
	"github.com/pahlevikun/applydiff/internal/runtime"
	"github.com/spf13/cobra"
)

func cmdInterview() *cobra.Command {
	c := &cobra.Command{
		Use:   "interview",
		Short: "Interview prep: kits, mocks, drills, and story bank",
	}
	c.AddCommand(
		cmdInterviewKit(),
		cmdInterviewMock(),
		cmdInterviewDrill(),
		cmdInterviewStory(),
		cmdInterviewScore(),
		cmdInterviewList(),
		cmdInterviewExport(),
	)
	return c
}

func cmdInterviewExport() *cobra.Command {
	var withPDF bool
	c := &cobra.Command{
		Use:   "export <job-key>",
		Short: "Export interview prep sheet to output/",
		Long: `Write output/<YYYY-MM-DD>-<company>-<role>-interview.md with loop, notes, rubric, stories, and weak spots.
With --pdf, also renders a PDF beside the markdown file using the same xelatex/Docker path as CV export.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			res, err := rootApp.ExportInterviewPrep(cmd.Context(), args[0], withPDF)
			exitOn(err)
			exitOn(emitOrJSON(res, func() string {
				if res.PDF != "" {
					return res.Markdown + "\n" + res.PDF
				}
				return res.Markdown
			}))
		},
	}
	c.Flags().BoolVar(&withPDF, "pdf", false, "also compile PDF via xelatex or Docker texlive")
	return c
}

func cmdInterviewKit() *cobra.Command {
	c := &cobra.Command{
		Use:   "kit <job-key>",
		Short: "Generate and store an interview kit for a job",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			kit, err := rootApp.GenerateInterviewKit(cmd.Context(), args[0])
			exitOn(err)
			exitOn(emitOrJSON(kit, func() string {
				return fmt.Sprintf("kit %s for %s", kit.ID, kit.ApplicationID)
			}))
		},
	}
	return c
}

func cmdInterviewMock() *cobra.Command {
	var mockType string
	c := &cobra.Command{
		Use:   "mock <job-key>",
		Short: "Start a mock interview session",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if mockType == "" {
				exitOn(fmt.Errorf("interview mock: --type required"))
			}
			sess, steps, err := rootApp.StartInterviewMock(cmd.Context(), args[0], mockType)
			exitOn(err)
			payload := runtime.MockSessionJSON{Session: *sess, Steps: steps}
			exitOn(emitOrJSON(payload, func() string {
				return fmt.Sprintf("session %s type=%s steps=%d", sess.ID, sess.Type, len(steps))
			}))
		},
	}
	c.Flags().StringVar(&mockType, "type", "", "system-design, coding, behavioral, business-case, or product-case")
	return c
}

func cmdInterviewDrill() *cobra.Command {
	var minutes int
	c := &cobra.Command{
		Use:   "drill <topic>",
		Short: "Run a short topic drill",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			note, drill, err := rootApp.InterviewDrill(cmd.Context(), args[0], minutes)
			exitOn(err)
			payload := runtime.DrillJSON{Topic: *note, Drill: *drill}
			exitOn(emitOrJSON(payload, func() string {
				return fmt.Sprintf("%s (%d min)\n%s\n\n%s", note.Title, drill.Minutes, note.Body, drill.Prompt)
			}))
		},
	}
	c.Flags().IntVar(&minutes, "minutes", 0, "override drill length in minutes")
	return c
}

func cmdInterviewStory() *cobra.Command {
	story := &cobra.Command{
		Use:   "story",
		Short: "Manage STAR stories",
	}
	story.AddCommand(cmdInterviewStoryAdd(), cmdInterviewStoryList())
	return story
}

func cmdInterviewStoryAdd() *cobra.Command {
	var category, situation, task, action, result, learning, evidence string
	c := &cobra.Command{
		Use:   "add <job-key>",
		Short: "Add a grounded STAR story",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			evidenceJSON := runtime.EncodeEvidenceJSON(nil)
			if evidence != "" {
				parts := strings.Split(evidence, ",")
				for i := range parts {
					parts[i] = strings.TrimSpace(parts[i])
				}
				evidenceJSON = runtime.EncodeEvidenceJSON(parts)
			}
			st := interview.Story{
				Category:     category,
				Situation:    situation,
				Task:         task,
				Action:       action,
				Result:       result,
				Learning:     learning,
				EvidenceJSON: evidenceJSON,
			}
			saved, err := rootApp.AddInterviewStory(cmd.Context(), args[0], st)
			exitOn(err)
			exitOn(emitOrJSON(saved, func() string {
				return fmt.Sprintf("story %s category=%s", saved.ID, saved.Category)
			}))
		},
	}
	c.Flags().StringVar(&category, "category", "", "story category")
	c.Flags().StringVar(&situation, "situation", "", "situation")
	c.Flags().StringVar(&task, "task", "", "task")
	c.Flags().StringVar(&action, "action", "", "action")
	c.Flags().StringVar(&result, "result", "", "result")
	c.Flags().StringVar(&learning, "learning", "", "learning")
	c.Flags().StringVar(&evidence, "evidence", "", "comma-separated fact keys")
	_ = c.MarkFlagRequired("category")
	_ = c.MarkFlagRequired("situation")
	_ = c.MarkFlagRequired("task")
	_ = c.MarkFlagRequired("action")
	_ = c.MarkFlagRequired("result")
	_ = c.MarkFlagRequired("learning")
	return c
}

func cmdInterviewStoryList() *cobra.Command {
	c := &cobra.Command{
		Use:   "list <job-key>",
		Short: "List STAR stories for a job",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			stories, err := rootApp.ListInterviewStories(cmd.Context(), args[0])
			exitOn(err)
			exitOn(emitOrJSON(stories, func() string {
				var b strings.Builder
				for _, st := range stories {
					b.WriteString(fmt.Sprintf("%s [%s]\n", st.ID, st.Category))
				}
				return strings.TrimRight(b.String(), "\n")
			}))
		},
	}
	return c
}

func cmdInterviewScore() *cobra.Command {
	var communication, problemSolving, correctness, depth, tradeoffs int
	var dimensions string
	c := &cobra.Command{
		Use:   "score <session-id>",
		Short: "Record rubric scores for a mock session",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			scores := map[string]int{}
			if dimensions != "" {
				parsed, err := runtime.ParseDimensionScores(dimensions)
				exitOn(err)
				for k, v := range parsed {
					scores[k] = v
				}
			}
			if communication > 0 {
				scores["communication"] = communication
			}
			if problemSolving > 0 {
				scores["problem-solving"] = problemSolving
			}
			if correctness > 0 {
				scores["correctness"] = correctness
			}
			if depth > 0 {
				scores["depth"] = depth
			}
			if tradeoffs > 0 {
				scores["tradeoffs"] = tradeoffs
			}
			if len(scores) == 0 {
				exitOn(fmt.Errorf("interview score: provide --dimension or named score flags"))
			}
			out, err := rootApp.ScoreInterviewSession(cmd.Context(), args[0], scores)
			exitOn(err)
			exitOn(emitOrJSON(out, func() string {
				return fmt.Sprintf("recorded %d rubric scores for %s", len(out), args[0])
			}))
		},
	}
	c.Flags().IntVar(&communication, "communication", 0, "communication score 1-5")
	c.Flags().IntVar(&problemSolving, "problem-solving", 0, "problem-solving score 1-5")
	c.Flags().IntVar(&correctness, "correctness", 0, "correctness score 1-5")
	c.Flags().IntVar(&depth, "depth", 0, "depth score 1-5")
	c.Flags().IntVar(&tradeoffs, "tradeoffs", 0, "tradeoffs score 1-5")
	c.Flags().StringVar(&dimensions, "dimension", "", "comma-separated dimension=score pairs")
	return c
}

func cmdInterviewList() *cobra.Command {
	c := &cobra.Command{
		Use:   "list <job-key>",
		Short: "List interview kits, sessions, and stories for a job",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			sum, err := rootApp.ListInterviewSummary(cmd.Context(), args[0])
			exitOn(err)
			exitOn(emitOrJSON(sum, func() string {
				return fmt.Sprintf("kits=%d sessions=%d stories=%d",
					len(sum.Kits), len(sum.Sessions), len(sum.Stories))
			}))
		},
	}
	return c
}
