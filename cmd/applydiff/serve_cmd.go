package main

import (
	"fmt"
	"os"

	"github.com/pahlevikun/applydiff/internal/runtime"
	"github.com/spf13/cobra"
)

func cmdServe() *cobra.Command {
	var port string
	var desktop bool
	c := &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP API (and static web when built)",
		Run: func(cmd *cobra.Command, args []string) {
			if desktop {
				_ = os.Setenv("APPLYDIFF_DESKTOP", "1")
				_ = runtime.EnsureLoginShellPATH()
			}
			if port == "" {
				port = "8080"
			}
			if desktop {
				fmt.Printf("LISTENING http://127.0.0.1:%s\n", port)
			}
			runHTTPServer(port, rootApp)
		},
	}
	c.Flags().StringVar(&port, "port", "8080", "listen port")
	c.Flags().BoolVar(&desktop, "desktop", false, "desktop sidecar mode (login-shell PATH, LISTENING line)")
	return c
}
