package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pahlevikun/applydiff/internal/present"
	"github.com/pahlevikun/applydiff/internal/runtime"
	"github.com/pahlevikun/applydiff/internal/theme"
	"github.com/spf13/cobra"
)

func cmdScan() *cobra.Command {
	var source, preset string
	c := &cobra.Command{
		Use:   "scan",
		Short: "Pull jobs from configured sources",
		Run: func(cmd *cobra.Command, args []string) {
			_ = source
			_ = preset
			res, err := rootApp.Scan(cmd.Context(), globalDryRun)
			exitOn(err)
			exitOn(emitOrJSON(res, func() string {
				return fmt.Sprintf("fetched=%d stored=%d skipped=%d", res.Fetched, res.Stored, res.Skipped)
			}))
		},
	}
	c.Flags().StringVar(&source, "source", "", "limit to one source id")
	c.Flags().StringVar(&preset, "preset", "", "apply filter preset before scan")
	return c
}

func cmdJobs() *cobra.Command {
	c := &cobra.Command{
		Use:   "jobs",
		Short: "List jobs with filter scores",
		Run: func(cmd *cobra.Command, args []string) {
			jobs, err := rootApp.ListJobs(cmd.Context())
			exitOn(err)
			exitOn(emitOrJSON(jobs, func() string {
				var b strings.Builder
				for _, j := range jobs {
					b.WriteString(fmt.Sprintf("%s  %s @ %s  score=%d keep=%v\n",
						j.Job.ID, j.Job.Title, j.Job.Company, j.Filter.Score, j.Filter.Keep))
				}
				return strings.TrimRight(b.String(), "\n")
			}))
		},
	}
	return c
}

func cmdShow() *cobra.Command {
	c := &cobra.Command{
		Use:   "show <job-key>",
		Short: "Show job detail and match reasons",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			detail, err := rootApp.ShowJob(cmd.Context(), args[0])
			exitOn(err)
			exitOn(emitOrJSON(detail, func() string {
				var b strings.Builder
				b.WriteString(fmt.Sprintf("%s — %s @ %s\n", detail.Job.ID, detail.Job.Title, detail.Job.Company))
				if detail.Match != nil {
					b.WriteString(fmt.Sprintf("match total: %d\n", detail.Match.Total))
					for _, r := range detail.Match.Requirements {
						b.WriteString(fmt.Sprintf("  [%s] %s\n", r.Status, r.Phrase))
					}
				}
				if detail.Job.Description != "" {
					b.WriteString("\n" + detail.Job.Description)
				}
				return b.String()
			}))
		},
	}
	return c
}

func cmdMatch() *cobra.Command {
	c := &cobra.Command{
		Use:   "match <job-key>",
		Short: "Compute match score vs master CV",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			res, err := rootApp.ComputeMatch(cmd.Context(), args[0], globalDryRun)
			if err != nil {
				if globalJSON && res != nil {
					_ = emitJSON(res)
				}
				exitOn(err)
			}
			exitOn(emitOrJSON(res, func() string {
				return fmt.Sprintf("total=%d job=%s", res.Total, res.JobID)
			}))
		},
	}
	return c
}

func cmdTailor() *cobra.Command {
	c := &cobra.Command{
		Use:   "tailor <job-key>",
		Short: "Generate tailoring hunks (LLM)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			prop, err := rootApp.Tailor(cmd.Context(), args[0], globalDryRun)
			exitOn(err)
			exitOn(emitOrJSON(prop, func() string {
				return fmt.Sprintf("%d hunks proposed for %s", len(prop.Hunks), args[0])
			}))
		},
	}
	return c
}

func cmdDiff() *cobra.Command {
	color := "auto"
	c := &cobra.Command{
		Use:   "diff <job-key>",
		Short: "Show unified diff per hunk",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			prop, err := rootApp.LoadProposal(cmd.Context(), args[0])
			exitOn(err)
			text := present.FormatUnified(prop.Hunks)
			useColor := color == "always" || (color == "auto" && isTTY(os.Stdout))
			if useColor {
				add := ansi(theme.Color("add"))
				del := ansi(theme.Color("del"))
				text = present.ColorizeUnified(text, add, del, true)
			}
			if globalJSON {
				exitOn(emitJSON(prop))
				return
			}
			fmt.Println(text)
		},
	}
	c.Flags().StringVar(&color, "color", "auto", "color output: auto, always, never")
	return c
}

func cmdApprove() *cobra.Command {
	var hunkStr string
	var all bool
	c := &cobra.Command{
		Use:   "approve <job-key>",
		Short: "Accept diff hunks",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var indices []int
			if all {
				indices = nil
			} else if hunkStr != "" {
				for _, part := range strings.Split(hunkStr, ",") {
					part = strings.TrimSpace(part)
					if part == "" {
						continue
					}
					var n int
					if _, err := fmt.Sscanf(part, "%d", &n); err != nil {
						exitOn(fmt.Errorf("invalid hunk: %s", part))
					}
					indices = append(indices, n)
				}
			} else {
				exitOn(fmt.Errorf("approve: specify --hunk N[,N] or --all"))
			}
			prop, err := rootApp.ApproveHunks(cmd.Context(), args[0], indices, all, globalDryRun)
			exitOn(err)
			exitOn(emitOrJSON(prop, func() string {
				return fmt.Sprintf("approved hunks for %s", args[0])
			}))
		},
	}
	c.Flags().StringVar(&hunkStr, "hunk", "", "hunk numbers to approve (1-based, comma-separated)")
	c.Flags().BoolVar(&all, "all", false, "approve all hunks")
	return c
}

func cmdReject() *cobra.Command {
	var hunkStr string
	c := &cobra.Command{
		Use:   "reject <job-key>",
		Short: "Reject diff hunks",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if hunkStr == "" {
				exitOn(fmt.Errorf("reject: --hunk N required"))
			}
			var indices []int
			for _, part := range strings.Split(hunkStr, ",") {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				var n int
				if _, err := fmt.Sscanf(part, "%d", &n); err != nil {
					exitOn(fmt.Errorf("invalid hunk: %s", part))
				}
				indices = append(indices, n)
			}
			prop, err := rootApp.RejectHunks(cmd.Context(), args[0], indices, globalDryRun)
			exitOn(err)
			exitOn(emitOrJSON(prop, func() string {
				return fmt.Sprintf("rejected hunks for %s", args[0])
			}))
		},
	}
	c.Flags().StringVar(&hunkStr, "hunk", "", "hunk numbers to reject")
	return c
}

func cmdBuild() *cobra.Command {
	c := &cobra.Command{
		Use:   "build <job-key>",
		Short: "Compile LaTeX PDF for approved branch",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			res, err := rootApp.BuildPDF(cmd.Context(), args[0], globalDryRun)
			exitOn(err)
			exitOn(emitOrJSON(res, func() string {
				if res.CVPDF != "" {
					return res.CVPDF
				}
				return res.Dir
			}))
		},
	}
	return c
}

func cmdApply() *cobra.Command {
	mode := "manual"
	c := &cobra.Command{
		Use:   "apply <job-key>",
		Short: "Record application submission (never auto-submits)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			res, err := rootApp.Apply(cmd.Context(), args[0], runtime.ApplyMode(mode), globalDryRun)
			exitOn(err)
			exitOn(emitOrJSON(res, func() string {
				return fmt.Sprintf("%s — %s (%s)", res.JobID, res.URL, res.Message)
			}))
		},
	}
	c.Flags().StringVar(&mode, "mode", "manual", "manual, assisted, or auto (auto disabled)")
	return c
}

func cmdExport() *cobra.Command {
	c := &cobra.Command{
		Use:   "export <file>",
		Short: "Export SQLite database to a file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			exitOn(rootApp.ExportDB(cmd.Context(), args[0], globalDryRun))
			if !globalJSON {
				fmt.Printf("exported -> %s\n", args[0])
			} else {
				exitOn(emitJSON(map[string]string{"path": args[0]}))
			}
		},
	}
	return c
}

func cmdImport() *cobra.Command {
	c := &cobra.Command{
		Use:   "import <file>",
		Short: "Import SQLite database from a file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			exitOn(rootApp.ImportDB(cmd.Context(), args[0], globalDryRun))
			if !globalJSON {
				fmt.Printf("imported %s\n", args[0])
			} else {
				exitOn(emitJSON(map[string]string{"path": args[0]}))
			}
		},
	}
	return c
}

func ansi(hex string) string {
	if len(hex) != 7 || hex[0] != '#' {
		return ""
	}
	var r, g, b int
	_, _ = fmt.Sscanf(hex[1:], "%02x%02x%02x", &r, &g, &b)
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

func isTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
