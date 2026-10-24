package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func cmdHunk() *cobra.Command {
	c := &cobra.Command{
		Use:   "hunk",
		Short: "Per-hunk review decisions",
	}
	c.AddCommand(cmdHunkDecide())
	return c
}

func cmdHunkDecide() *cobra.Command {
	var index int
	var decision string
	var text string
	c := &cobra.Command{
		Use:   "decide <job-key>",
		Short: "Accept, reject, edit, or skip one hunk",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if index < 1 {
				exitOn(fmt.Errorf("hunk decide: --index required (1-based)"))
			}
			if decision == "" {
				exitOn(fmt.Errorf("hunk decide: --decision required"))
			}
			sum, err := rootApp.SetHunkDecision(cmd.Context(), args[0], index, decision, text, globalDryRun, actor())
			exitOn(err)
			payload := map[string]int{"decided": sum.Decided, "total": sum.Total}
			exitOn(emitOrJSON(payload, func() string {
				return fmt.Sprintf("%d of %d decided for %s", sum.Decided, sum.Total, args[0])
			}))
		},
	}
	c.Flags().IntVar(&index, "index", 0, "1-based hunk index")
	c.Flags().StringVar(&decision, "decision", "", "accept, reject, edit, or skip")
	c.Flags().StringVar(&text, "text", "", "edited text or skip reason")
	return c
}
