package store

import (
	"context"
	"testing"
)

func TestMigrateFreshAndIdempotent(t *testing.T) {
	ctx := context.Background()
	st, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	want, err := ListMigrationFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(want) == 0 {
		t.Fatal("expected embedded migrations")
	}

	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	applied, err := st.AppliedMigrations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != len(want) {
		t.Fatalf("applied %d migrations, want %d: %v", len(applied), len(want), applied)
	}
	for i := range want {
		if applied[i] != want[i] {
			t.Fatalf("order mismatch at %d: got %q want %q", i, applied[i], want[i])
		}
	}

	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	applied2, err := st.AppliedMigrations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied2) != len(applied) {
		t.Fatalf("re-migrate changed count: %d -> %d", len(applied), len(applied2))
	}
}
