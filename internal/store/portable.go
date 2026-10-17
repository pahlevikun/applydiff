package store

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func sqliteQuoteLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// ExportDB writes a consistent SQLite snapshot to destPath (replaces an existing file).
func (s *Store) ExportDB(ctx context.Context, destPath string) error {
	destPath = filepath.Clean(destPath)
	if err := os.Remove(destPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	q := fmt.Sprintf("VACUUM INTO %s", sqliteQuoteLiteral(destPath))
	if _, err := s.db.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("export db: %w", err)
	}
	return nil
}

// ImportDB replaces destPath with the SQLite file at srcPath and opens a new store.
func ImportDB(ctx context.Context, srcPath, destPath string) (*Store, error) {
	srcPath = filepath.Clean(srcPath)
	destPath = filepath.Clean(destPath)
	if err := validateSQLiteFile(srcPath); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil && filepath.Dir(destPath) != "." {
		return nil, err
	}
	in, err := os.Open(srcPath)
	if err != nil {
		return nil, err
	}
	defer in.Close()
	tmp := destPath + ".importing"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tmp)
		return nil, err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return nil, err
	}
	if err := os.Rename(tmp, destPath); err != nil {
		os.Remove(tmp)
		return nil, err
	}
	return New(destPath)
}

func validateSQLiteFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	hdr := make([]byte, 16)
	if _, err := io.ReadFull(f, hdr); err != nil {
		return fmt.Errorf("import db: not a sqlite file: %w", err)
	}
	if string(hdr[:15]) != "SQLite format 3" {
		return fmt.Errorf("import db: not a sqlite file")
	}
	return nil
}
