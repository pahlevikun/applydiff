package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"github.com/pahlevikun/applydiff/internal/export"
	"github.com/pahlevikun/applydiff/internal/store"
	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build <job-key>",
		Short: "Compile an approved branch CV to a PDF artifact",
		Long: `Build writes a single PDF under output/ using the job branch CV.

The file name is output/<YYYY-MM-DD>-<company>-<role>.pdf (company and role are slugified).
The output/ directory is created when missing. No per-application folders or state files are written.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(args[0])
		},
	}
	return cmd
}

func runBuild(jobKey string) error {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "applydiff.db"
	}
	st, err := store.New(dbPath)
	if err != nil {
		return err
	}
	defer st.Close()
	ctx := context.Background()
	if err := st.Migrate(ctx); err != nil {
		return err
	}

	j, err := st.GetJob(ctx, jobKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("job not found: %s", jobKey)
		}
		return err
	}
	branch, err := st.LoadBranch(ctx, jobKey)
	if err != nil {
		return fmt.Errorf("branch cv: %w", err)
	}
	facts, err := st.LoadFacts(ctx)
	if err != nil {
		facts = nil
	}
	outRoot := os.Getenv("OUTPUT_ROOT")
	if outRoot == "" {
		outRoot = "output"
	}
	pdfPath, snapshot, err := export.BuildPDF(export.Options{
		OutRoot: outRoot,
		Job:     *j,
		Doc:     branch,
		Facts:   facts,
	})
	if err != nil {
		return err
	}
	fmt.Println(pdfPath)
	_ = snapshot
	return nil
}
