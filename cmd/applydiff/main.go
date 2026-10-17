package main

import (
	"fmt"
	"os"

	"github.com/pahlevikun/applydiff/internal/runtime"
)

func main() {
	// Default: serve on PORT for backward compatibility when no subcommand is given.
	if len(os.Args) == 1 {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		ctx := rootCtx
		app, err := runtime.Open(ctx, runtime.Options{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		runHTTPServer(port, app)
		return
	}
	// Legacy flags before cobra subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "export-db":
			runExportDB(os.Args[2:])
			return
		case "import-db":
			runImportDB(os.Args[2:])
			return
		}
	}
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
