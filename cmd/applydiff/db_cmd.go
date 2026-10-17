package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pahlevikun/applydiff/internal/store"
	"github.com/spf13/cobra"
)

func newExportDBCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "export-db",
		Short: "Export the SQLite database to a file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if out == "" {
				return fmt.Errorf("export-db: -o <file> is required")
			}
			dbPath := os.Getenv("DATABASE_PATH")
			if dbPath == "" {
				dbPath = "applydiff.db"
			}
			st, err := store.New(dbPath)
			if err != nil {
				return err
			}
			defer st.Close()
			ctx := context.Background()
			if err := st.ExportDB(ctx, out); err != nil {
				return err
			}
			fmt.Printf("exported %s -> %s\n", dbPath, out)
			return nil
		},
	}
	cmd.Flags().StringVarP(&out, "o", "o", "", "output sqlite file")
	return cmd
}

func newImportDBCmd() *cobra.Command {
	var in string
	cmd := &cobra.Command{
		Use:   "import-db",
		Short: "Import a SQLite database file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if in == "" {
				return fmt.Errorf("import-db: -i <file> is required")
			}
			dbPath := os.Getenv("DATABASE_PATH")
			if dbPath == "" {
				dbPath = "applydiff.db"
			}
			ctx := context.Background()
			st, err := store.ImportDB(ctx, in, dbPath)
			if err != nil {
				return err
			}
			st.Close()
			fmt.Printf("imported %s -> %s\n", in, dbPath)
			return nil
		},
	}
	cmd.Flags().StringVarP(&in, "i", "i", "", "input sqlite file")
	return cmd
}
