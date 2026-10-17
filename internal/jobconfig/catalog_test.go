package jobconfig

import (
	"testing"
)

func TestSourceCatalogMergedInventory(t *testing.T) {
	cat := SourceCatalog()
	if len(cat) < 120 {
		t.Fatalf("expected merged catalog (career-ops + extras), got %d", len(cat))
	}
	total, live, planned, manual := CatalogStats()
	if total != len(cat) {
		t.Fatalf("stats total mismatch: %d vs %d", total, len(cat))
	}
	if live != 7 {
		t.Fatalf("expected 7 live scanners, got %d", live)
	}
	if manual < 1 {
		t.Fatalf("expected manual row")
	}
	if planned < 100 {
		t.Fatalf("expected 100+ planned career-ops rows, got %d", planned)
	}
	ids := map[string]bool{}
	for _, e := range cat {
		if ids[e.ID] {
			t.Fatalf("duplicate catalog id: %s", e.ID)
		}
		ids[e.ID] = true
	}
	for _, must := range []string{"freehire", "jobnet", "tokyodev", "linkedin_alerts", "glassdoor", "ashby", "lever", "fourdayweek"} {
		if !ids[must] {
			t.Fatalf("missing expected id %s", must)
		}
	}
}
