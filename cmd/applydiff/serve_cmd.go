package main

import (
	"github.com/spf13/cobra"
)

func cmdServe() *cobra.Command {
	var port string
	c := &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP API (and static web when built)",
		Run: func(cmd *cobra.Command, args []string) {
			if port == "" {
				port = "8787"
			}
			runHTTPServer(port, rootApp)
		},
	}
	c.Flags().StringVar(&port, "port", "8787", "listen port")
	return c
}
