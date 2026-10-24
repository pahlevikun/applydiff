package main

import (
	"time"

	"github.com/pahlevikun/applydiff/internal/daemon"
	"github.com/spf13/cobra"
)

func cmdDaemon() *cobra.Command {
	var interval time.Duration
	var drafts int
	var once bool
	var notify bool
	c := &cobra.Command{
		Use:   "daemon",
		Short: "Run scheduled scan, match, and optional draft loop",
		Run: func(cmd *cobra.Command, args []string) {
			opt := daemon.Options{
				Interval:        interval,
				MaxDraftsPerDay: drafts,
				Once:            once,
				DryRun:          globalDryRun,
				Notify:          notify,
				DataDir:         rootApp.Paths.Home,
			}
			exitOn(daemon.Run(cmd.Context(), rootApp, opt))
		},
	}
	c.Flags().DurationVar(&interval, "interval", 6*time.Hour, "time between daemon ticks")
	c.Flags().IntVar(&drafts, "drafts", 10, "max auto-drafts per rolling 24h")
	c.Flags().BoolVar(&once, "once", false, "run one tick then exit")
	c.Flags().BoolVar(&notify, "notify", false, "send notifications after a tick (off by default)")
	return c
}
