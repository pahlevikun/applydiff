package main

import (
	"fmt"
	"strings"

	"github.com/pahlevikun/applydiff/internal/inbox"
	"github.com/spf13/cobra"
)

func cmdInbox() *cobra.Command {
	var state string
	var limit int
	c := &cobra.Command{
		Use:   "inbox",
		Short: "List the decision queue",
		Run: func(cmd *cobra.Command, args []string) {
			res, err := inbox.List(cmd.Context(), rootApp.Store, inbox.ListOptions{
				State: state,
				Limit: limit,
			})
			exitOn(err)
			exitOn(emitOrJSON(res, func() string { return formatInboxTable(res) }))
		},
	}
	c.Flags().StringVar(&state, "state", "", "filter: needs-decision, in-progress, sent, interview")
	c.Flags().IntVar(&limit, "limit", 0, "max rows (0 = all)")
	return c
}

func formatInboxTable(res inbox.Result) string {
	var b strings.Builder
	for _, row := range res.Data {
		b.WriteString(fmt.Sprintf("%-16s  %-28s  %s @ %s  %s\n",
			row.State, row.JobKey, row.Title, row.Company, row.Reason))
	}
	b.WriteString(fmt.Sprintf("\ncounts: needs_decision=%d in_progress=%d sent=%d interview=%d\n",
		res.Meta.Counts.NeedsDecision, res.Meta.Counts.InProgress,
		res.Meta.Counts.Sent, res.Meta.Counts.Interview))
	return strings.TrimRight(b.String(), "\n")
}
