package llm

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type cliRunner struct {
	command string
}

func NewCLIRunner(cfg Config) (Runner, error) {
	cmd := strings.TrimSpace(cfg.CLI)
	if cmd == "" {
		return nil, fmt.Errorf("llm: LLM_CLI required for cli provider")
	}
	return &cliRunner{command: cmd}, nil
}

func (c *cliRunner) Complete(ctx context.Context, prompt string) (string, error) {
	parts := strings.Fields(c.command)
	if len(parts) == 0 {
		return "", fmt.Errorf("llm: empty CLI command")
	}
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Stdin = bytes.NewBufferString(prompt)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("llm cli: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
