package store

import (
	"context"
	_ "embed"
)

//go:embed migrations/001_jobs.sql
var migration001 string

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, migration001)
	return err
}
