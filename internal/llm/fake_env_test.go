package llm

import (
	"testing"
)

func TestFakeResponseFromEnv(t *testing.T) {
	t.Setenv(envFakeResponse, `{"hunks":[{"field_id":"x"}]}`)
	r, err := NewRunner(Config{Provider: "fake"})
	if err != nil {
		t.Fatal(err)
	}
	f, ok := r.(FakeRunner)
	if !ok {
		t.Fatalf("expected FakeRunner, got %T", r)
	}
	if f.Response != `{"hunks":[{"field_id":"x"}]}` {
		t.Fatalf("unexpected response %q", f.Response)
	}
}
