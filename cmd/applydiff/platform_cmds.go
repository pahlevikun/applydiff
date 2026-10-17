package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pahlevikun/applydiff/internal/runtime"
	"github.com/pahlevikun/applydiff/internal/tailor"
	"github.com/spf13/cobra"
)

func cmdAdd() *cobra.Command {
	var title, company, url, source string
	c := &cobra.Command{
		Use:   "add",
		Short: "Add a pasted job description (manual source)",
		Run: func(cmd *cobra.Command, args []string) {
			desc, err := readJobDescription(cmd)
			exitOn(err)
			j, err := rootApp.AddJob(cmd.Context(), runtime.AddJobInput{
				Title: title, Company: company, URL: url, Description: desc, Source: source,
			}, runtime.CurrentActor())
			exitOn(err)
			exitOn(emitOrJSON(j, func() string {
				return fmt.Sprintf("%s — %s @ %s", j.ID, j.Title, j.Company)
			}))
		},
	}
	c.Flags().StringVar(&title, "title", "", "job title")
	c.Flags().StringVar(&company, "company", "", "company name")
	c.Flags().StringVar(&url, "url", "", "job posting URL")
	c.Flags().StringVar(&source, "source", "manual", "source id")
	return c
}

func readJobDescription(cmd *cobra.Command) (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	s := strings.TrimSpace(string(data))
	if s == "" {
		return "", fmt.Errorf("add: paste job description on stdin")
	}
	return s, nil
}

func cmdPropose() *cobra.Command {
	var hunksFile string
	c := &cobra.Command{
		Use:   "propose <job-key>",
		Short: "Validate and save host-drafted hunks (JSON on stdin)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, err := readRawHunks(hunksFile)
			exitOn(err)
			prop, err := rootApp.ProposeHunks(cmd.Context(), args[0], raw, runtime.CurrentActor())
			exitOn(err)
			exitOn(emitOrJSON(prop, func() string {
				return fmt.Sprintf("%d hunks proposed", len(prop.Hunks))
			}))
		},
	}
	c.Flags().StringVar(&hunksFile, "hunks", "-", "hunks JSON file or - for stdin")
	return c
}

func readRawHunks(path string) ([]tailor.RawHunk, error) {
	var r io.Reader = os.Stdin
	if path != "" && path != "-" {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	}
	var env struct {
		Hunks []tailor.RawHunk `json:"hunks"`
	}
	if err := json.NewDecoder(r).Decode(&env); err != nil {
		return nil, err
	}
	return env.Hunks, nil
}

func cmdRuns() *cobra.Command {
	c := &cobra.Command{Use: "runs", Short: "Async run status"}
	c.AddCommand(cmdRunsGet())
	return c
}

func cmdRunsGet() *cobra.Command {
	c := &cobra.Command{
		Use:   "get <run-id>",
		Short: "Fetch async run result",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			run, err := rootApp.Runs.Get(cmd.Context(), args[0])
			exitOn(err)
			exitOn(emitOrJSON(run, func() string {
				return fmt.Sprintf("%s %s %s", run.ID, run.Status, run.Kind)
			}))
		},
	}
	return c
}
