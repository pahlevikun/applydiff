package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pahlevikun/applydiff/internal/paths"
	"github.com/pahlevikun/applydiff/internal/store"
)

func legacyDBPath() string {
	if dbPath := os.Getenv("DATABASE_PATH"); dbPath != "" {
		return dbPath
	}
	p, err := paths.Resolve()
	if err != nil {
		return "applydiff.db"
	}
	return p.DB
}

func runExportDB(args []string) {
	fs := flag.NewFlagSet("export-db", flag.ExitOnError)
	out := fs.String("o", "", "output sqlite file (required)")
	_ = fs.Parse(args)
	if *out == "" {
		fmt.Fprintln(os.Stderr, "export-db: -o <file> is required")
		os.Exit(2)
	}
	dbPath := legacyDBPath()
	st, err := store.New(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	if err := st.ExportDB(ctx, *out); err != nil {
		log.Fatal(err)
	}
	log.Printf("exported %s -> %s", dbPath, *out)
}

func runImportDB(args []string) {
	fs := flag.NewFlagSet("import-db", flag.ExitOnError)
	in := fs.String("i", "", "input sqlite file (required)")
	_ = fs.Parse(args)
	if *in == "" {
		fmt.Fprintln(os.Stderr, "import-db: -i <file> is required")
		os.Exit(2)
	}
	dbPath := legacyDBPath()
	ctx := context.Background()
	st, err := store.ImportDB(ctx, *in, dbPath)
	if err != nil {
		log.Fatal(err)
	}
	st.Close()
	log.Printf("imported %s -> %s", *in, dbPath)
}
