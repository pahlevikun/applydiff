package llm

import (
	"context"
	"fmt"
	"log"
)

// Runner completes a single prompt and returns the model text (expected JSON for tailor).
type Runner interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

// NewRunner builds the configured backend.
func NewRunner(cfg Config) (Runner, error) {
	switch cfg.Provider {
	case "api":
		return NewAPIRunner(cfg)
	case "cli":
		return NewCLIRunner(cfg)
	case "fake":
		return FakeRunner{Response: fakeResponseFromEnv()}, nil
	default:
		return nil, fmt.Errorf("llm: unknown provider %q", cfg.Provider)
	}
}

// NewRunnerFromEnv resolves env config and returns a runner (fake when no backend is available).
func NewRunnerFromEnv() Runner {
	cfg := ConfigFromEnv()
	runner, err := NewRunner(cfg)
	if err != nil {
		log.Printf("llm: using fake runner (%v)", err)
		return FakeRunner{Response: fakeResponseFromEnv()}
	}
	if cfg.Provider == "fake" {
		log.Printf("llm: no LLM_API_KEY or agent CLI configured; using fake runner (set LLM_CLI or install claude/codex/cursor-agent/hermes)")
	}
	return runner
}
