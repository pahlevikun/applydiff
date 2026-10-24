package job

import "testing"

func TestRoleKeyCollapsesVariants(t *testing.T) {
	cases := []struct {
		companyA, titleA, companyB, titleB string
	}{
		{"Acme Inc", "Senior Backend Engineer, Krakow", "ACME", "senior backend engineer"},
		{"Foo  Bar", "Go Developer (Remote)", "Foo Bar", "go developer remote"},
		{"Acme Ltd", "Engineer - EU", "Acme", "Engineer"},
	}
	for _, tc := range cases {
		ka := RoleKey(tc.companyA, tc.titleA)
		kb := RoleKey(tc.companyB, tc.titleB)
		if ka == "" || kb == "" {
			t.Fatalf("unexpected empty key for %+v", tc)
		}
		if ka != kb {
			t.Fatalf("expected same key for %+v: %q vs %q", tc, ka, kb)
		}
	}
}

func TestRoleKeyDoesNotCollapseDifferentRoles(t *testing.T) {
	goKey := RoleKey("Acme", "Go Developer")
	rustKey := RoleKey("Acme", "Rust Developer")
	if goKey == "" || rustKey == "" {
		t.Fatal("expected non-empty keys")
	}
	if goKey == rustKey {
		t.Fatalf("Go and Rust roles must not share role_key: %q", goKey)
	}
}
