package store

import (
	"context"
	"testing"

	"github.com/pahlevikun/applydiff/internal/filter"
)

func TestFilterPresetSaveApplyRoundTrip(t *testing.T) {
	ctx := context.Background()
	st, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	spec := filter.Spec{
		Arrangement: &filter.CategoryRule{Hard: true, Values: []string{"remote"}},
		Regions:     &filter.CategoryRule{Hard: true, Values: []string{"SEA"}},
	}
	p, err := st.SaveFilterPreset(ctx, "Remote SEA", spec)
	if err != nil {
		t.Fatal(err)
	}
	list, err := st.ListFilterPresets(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("list %v err %v", list, err)
	}
	got, err := st.GetFilterPreset(ctx, p.ID)
	if err != nil || got.Name != "Remote SEA" {
		t.Fatalf("get %+v err %v", got, err)
	}
	if err := st.SetActiveFilterPresetID(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	active, err := st.GetActiveFilterPresetID(ctx)
	if err != nil || active != p.ID {
		t.Fatalf("active %q", active)
	}
	if err := st.DeleteFilterPreset(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	active, _ = st.GetActiveFilterPresetID(ctx)
	if active != "" {
		t.Fatalf("expected cleared active, got %q", active)
	}
}
