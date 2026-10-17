package store

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", sqliteDSN(path))
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func sqliteDSN(path string) string {
	if path == ":memory:" || strings.HasPrefix(path, "file:") {
		return path
	}
	abs := path
	if a, err := filepath.Abs(path); err == nil {
		abs = a
	}
	abs = filepath.ToSlash(abs)
	return fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", abs)
}

func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
