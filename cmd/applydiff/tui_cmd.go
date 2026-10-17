package main

import (
	"os"

	"github.com/pahlevikun/applydiff/cmd/applydiff/tui"
	"github.com/spf13/cobra"
)

func cmdTUI() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Interactive terminal UI",
		Run: func(cmd *cobra.Command, args []string) {
			if err := tui.Run(rootApp); err != nil {
				os.Exit(1)
			}
		},
	}
}
