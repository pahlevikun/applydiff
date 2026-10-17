package runtime

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pahlevikun/applydiff/internal/cv"
	cvexport "github.com/pahlevikun/applydiff/internal/cv/export"
)

// MasterCVSource indicates where LoadMasterCV read the document from.
type MasterCVSource string

const (
	MasterCVFromDB   MasterCVSource = "database"
	MasterCVFromFile MasterCVSource = "file"
)

// MasterCVLoadResult is the structured master plus provenance for UIs.
type MasterCVLoadResult struct {
	Document   *cv.Document
	Source     MasterCVSource
	MasterJSON string
	MasterTeX  string
}

// LoadMasterCV returns the master from the database, or cv/master_cv.json when empty.
func (a *App) LoadMasterCV(ctx context.Context) (*MasterCVLoadResult, error) {
	if a == nil {
		return nil, fmt.Errorf("app not initialized")
	}
	res := MasterCVLoadResult{
		MasterJSON: a.Paths.MasterJSON,
		MasterTeX:  a.Paths.MasterTeX,
	}
	doc, err := a.Store.LoadMaster(ctx)
	if err == nil {
		res.Document = doc
		res.Source = MasterCVFromDB
		return &res, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	doc, err = cv.LoadFile(a.Paths.MasterJSON)
	if err != nil {
		return nil, fmt.Errorf("no master cv in database and %s: %w", a.Paths.MasterJSON, err)
	}
	res.Document = doc
	res.Source = MasterCVFromFile
	return &res, nil
}

// SaveMasterCV persists to the database and canonical JSON / TeX sources.
func (a *App) SaveMasterCV(ctx context.Context, doc *cv.Document) error {
	if a == nil {
		return fmt.Errorf("app not initialized")
	}
	if err := cv.Validate(doc); err != nil {
		return err
	}
	if err := a.Store.SaveMaster(ctx, doc); err != nil {
		return err
	}
	if err := cv.WriteMasterJSON(a.Paths.MasterJSON, doc); err != nil {
		return fmt.Errorf("write master json: %w", err)
	}
	tex := cvexport.ToTeX(doc)
	if err := os.MkdirAll(filepath.Dir(a.Paths.MasterTeX), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(a.Paths.MasterTeX, []byte(tex), 0o644); err != nil {
		return fmt.Errorf("write master tex: %w", err)
	}
	return nil
}
