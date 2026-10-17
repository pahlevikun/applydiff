package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pahlevikun/applydiff/internal/runtime"
	"github.com/spf13/cobra"
)

var (
	globalJSON    bool
	globalDryRun  bool
	rootApp       *runtime.App
	rootCtx       = context.Background()
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "applydiff",
		Short: "ApplyDiff — scan, match, tailor, and approve CV diffs",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if rootApp != nil {
				return nil
			}
			app, err := runtime.Open(rootCtx, runtime.Options{})
			if err != nil {
				return err
			}
			rootApp = app
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if rootApp != nil {
				_ = rootApp.Close()
				rootApp = nil
			}
		},
	}
	root.PersistentFlags().BoolVar(&globalJSON, "json", false, "emit JSON on stdout")
	root.PersistentFlags().BoolVar(&globalDryRun, "dry-run", false, "do not write changes")

	root.AddCommand(
		cmdScan(),
		cmdJobs(),
		cmdShow(),
		cmdMatch(),
		cmdTailor(),
		cmdDiff(),
		cmdApprove(),
		cmdReject(),
		cmdBuild(),
		cmdApply(),
		cmdExport(),
		cmdImport(),
		cmdServe(),
		cmdTUI(),
	)
	// Legacy entrypoints
	root.AddCommand(cmdExportDBLegacy(), cmdImportDBLegacy())
	return root
}

func emitJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func emitOrJSON(v interface{}, text func() string) error {
	if globalJSON {
		return emitJSON(v)
	}
	if text != nil {
		fmt.Println(text())
	}
	return nil
}

func exitOn(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(runtime.ExitCode(err))
}
