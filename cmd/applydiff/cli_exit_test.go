package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/pahlevikun/applydiff/internal/runtime"
)

func TestCLIHelpListsSubcommands(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, sub := range []string{
		"doctor", "init", "open",
		"scan", "jobs", "show", "match", "tailor", "diff", "approve", "reject",
		"build", "apply", "stats", "export", "import", "serve", "tui",
	} {
		if !containsWord(out, sub) {
			t.Fatalf("help missing subcommand %q:\n%s", sub, out)
		}
	}
}

func containsWord(hay, word string) bool {
	return bytes.Contains([]byte(hay), []byte(word))
}

func TestCLIExitCodeHelpers(t *testing.T) {
	if runtime.ExitCode(runtime.ErrLLM{Err: os.ErrInvalid}) != 4 {
		t.Fatal("llm exit")
	}
}
