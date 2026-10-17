package main

import (
	"github.com/spf13/cobra"
)

func cmdExportDBLegacy() *cobra.Command {
	return &cobra.Command{
		Use:    "export-db",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			runExportDB(args)
		},
	}
}

func cmdImportDBLegacy() *cobra.Command {
	return &cobra.Command{
		Use:    "import-db",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			runImportDB(args)
		},
	}
}
