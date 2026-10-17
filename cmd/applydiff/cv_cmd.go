package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func cmdCV() *cobra.Command {
	root := &cobra.Command{
		Use:   "cv",
		Short: "Manage the master CV",
	}
	root.AddCommand(cmdCVImport())
	return root
}

func cmdCVImport() *cobra.Command {
	c := &cobra.Command{
		Use:   "import <master_cv.json>",
		Short: "Import structured master CV (and sibling facts.json when present)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := rootApp.ImportMasterCV(cmd.Context(), args[0], globalDryRun)
			exitOn(err)
			if !globalJSON {
				fmt.Printf("imported master cv from %s\n", args[0])
			} else {
				exitOn(emitJSON(map[string]string{"status": "imported", "path": args[0]}))
			}
		},
	}
	return c
}
