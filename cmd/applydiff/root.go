package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "applydiff",
		Short: "Git-style approval for job applications",
	}
	root.AddCommand(
		newServeCmd(),
		newBuildCmd(),
		newTUICmd(),
		newExportDBCmd(),
		newImportDBCmd(),
	)
	root.SilenceUsage = true
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
