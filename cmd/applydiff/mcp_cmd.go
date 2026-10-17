package main

import (
	"fmt"

	"github.com/pahlevikun/applydiff/internal/mcpserver"
	"github.com/spf13/cobra"
)

func cmdMCP() *cobra.Command {
	c := &cobra.Command{
		Use:   "mcp",
		Short: "Run MCP server over stdio",
		Run: func(cmd *cobra.Command, args []string) {
			exitOn(mcpserver.Run(cmd.Context(), rootApp))
		},
	}
	c.AddCommand(cmdMCPInstall())
	return c
}

func cmdMCPInstall() *cobra.Command {
	c := &cobra.Command{
		Use:   "install <cursor|claude|codex>",
		Short: "Register applydiff mcp in a client config",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			exitOn(mcpserver.InstallMCP(args[0], ""))
			if !globalJSON {
				fmt.Printf("registered applydiff mcp for %s\n", args[0])
			}
		},
	}
	return c
}
