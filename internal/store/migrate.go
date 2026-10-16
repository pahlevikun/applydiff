package store

import (
	"context"
	_ "embed"
)

//go:embed migrations/001_jobs.sql
var migration001 string

//go:embed migrations/002_cv.sql
var migration002 string

//go:embed migrations/003_match.sql
var migration003 string

//go:embed migrations/004_tailor.sql
var migration004 string

func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, migration001); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, migration002); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, migration003); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, migration004)
	return err
}
