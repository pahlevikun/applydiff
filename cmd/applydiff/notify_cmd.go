package main

import (
	"github.com/pahlevikun/applydiff/internal/notify"
	"github.com/spf13/cobra"
)

func cmdNotify() *cobra.Command {
	c := &cobra.Command{
		Use:   "notify",
		Short: "Notification channels (test and status)",
	}
	c.AddCommand(cmdNotifyTest(), cmdNotifyStatus())
	return c
}

func cmdNotifyTest() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Send one test message via the configured notifier",
		Run: func(cmd *cobra.Command, args []string) {
			exitOn(notify.SendTest(cmd.Context()))
		},
	}
}

func cmdNotifyStatus() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show configured notification channels",
		Run: func(cmd *cobra.Command, args []string) {
			exitOn(emitOrJSON(notify.StatusLines(), func() string { return notify.FormatStatus() }))
		},
	}
}
