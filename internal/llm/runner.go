package llm

import (
	"context"
	"fmt"
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
	default:
		return nil, fmt.Errorf("llm: unknown provider %q", cfg.Provider)
	}
}
