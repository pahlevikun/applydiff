package tailor

import (
	"context"
	"testing"

	"github.com/pahlevikun/applydiff/internal/llm"
)

func TestFakeRunnerReturnsCannedHunks(t *testing.T) {
	canned := `{"hunks":[{"field_id":"summary","section":"summary","before":"Old","after":"New lead","rationale":"mirror JD","evidence":["fact1"]}]}`
	runner := llm.FakeRunner{Response: canned}
	out, err := runner.Complete(context.Background(), "prompt")
	if err != nil {
		t.Fatal(err)
	}
	hunks, err := ParseHunksJSON(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(hunks) != 1 || hunks[0].FieldID != "summary" {
		t.Fatalf("unexpected hunks: %+v", hunks)
	}
}
