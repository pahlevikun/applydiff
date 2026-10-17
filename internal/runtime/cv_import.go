package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pahlevikun/applydiff/internal/cv"
)

// ImportMasterCV loads a structured CV JSON file and optional sibling facts.json into the store.
func (a *App) ImportMasterCV(ctx context.Context, masterPath string, dryRun bool) error {
	if a == nil {
		return fmt.Errorf("app not initialized")
	}
	doc, err := cv.LoadFile(masterPath)
	if err != nil {
		return err
	}
	factsPath := filepath.Join(filepath.Dir(masterPath), "facts.json")
	var facts *cv.Facts
	if data, err := os.ReadFile(factsPath); err == nil {
		facts, err = cv.LoadFactsJSON(data)
		if err != nil {
			return fmt.Errorf("facts: %w", err)
		}
	}
	if dryRun {
		return nil
	}
	if err := a.Store.SaveMaster(ctx, doc); err != nil {
		return err
	}
	if facts != nil {
		if err := a.Store.SaveFacts(ctx, facts); err != nil {
			return err
		}
		a.Facts = facts
	}
	return nil
}
