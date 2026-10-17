package llm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigFromEnvProviderSelection(t *testing.T) {
	t.Setenv("LLM_PROVIDER", "")
	t.Setenv("LLM_CLI", "")
	t.Setenv("LLM_API_KEY", "")
	t.Setenv("PATH", t.TempDir())

	cfg := ConfigFromEnv()
	if cfg.Provider != "fake" {
		t.Fatalf("expected fake, got %q", cfg.Provider)
	}

	t.Setenv("LLM_API_KEY", "sk-test")
	cfg = ConfigFromEnv()
	if cfg.Provider != "api" {
		t.Fatalf("expected api, got %q", cfg.Provider)
	}

	t.Setenv("LLM_API_KEY", "")
	t.Setenv("LLM_CLI", "echo hi")
	cfg = ConfigFromEnv()
	if cfg.Provider != "cli" || cfg.CLI != "echo hi" {
		t.Fatalf("expected cli with LLM_CLI, got %+v", cfg)
	}

	t.Setenv("LLM_CLI", "")
	bin := t.TempDir()
	claudePath := filepath.Join(bin, "claude")
	if err := os.WriteFile(claudePath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	cfg = ConfigFromEnv()
	if cfg.Provider != "cli" || cfg.CLI != "claude -p" {
		t.Fatalf("expected auto-detected claude cli, got %+v", cfg)
	}

	t.Setenv("LLM_PROVIDER", "api")
	t.Setenv("LLM_API_KEY", "")
	cfg = ConfigFromEnv()
	if cfg.Provider != "api" {
		t.Fatalf("explicit provider should win, got %q", cfg.Provider)
	}
}

func TestNewRunnerFromEnv(t *testing.T) {
	t.Setenv("LLM_PROVIDER", "")
	t.Setenv("LLM_CLI", "")
	t.Setenv("LLM_API_KEY", "")
	t.Setenv("PATH", t.TempDir())

	r := NewRunnerFromEnv()
	if _, ok := r.(FakeRunner); !ok {
		t.Fatalf("expected FakeRunner, got %T", r)
	}

	t.Setenv("LLM_API_KEY", "sk-test")
	r = NewRunnerFromEnv()
	if _, ok := r.(*apiRunner); !ok {
		t.Fatalf("expected apiRunner, got %T", r)
	}

	t.Setenv("LLM_API_KEY", "")
	t.Setenv("LLM_CLI", "echo ok")
	r = NewRunnerFromEnv()
	if _, ok := r.(*cliRunner); !ok {
		t.Fatalf("expected cliRunner, got %T", r)
	}
}
