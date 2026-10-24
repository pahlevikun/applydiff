package mcpserver

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CLI runs applydiff subcommands and returns stdout (JSON) on success.
type CLI struct {
	Bin string
}

// ResolveCLI picks the binary used for MCP tool calls.
func ResolveCLI() string {
	if v := os.Getenv("APPLYDIFF_BIN"); v != "" {
		return v
	}
	exe, err := os.Executable()
	if err != nil {
		return "applydiff"
	}
	return exe
}

func (c CLI) run(ctx context.Context, args ...string) (string, error) {
	if c.Bin == "" {
		c.Bin = ResolveCLI()
	}
	cmd := exec.CommandContext(ctx, c.Bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s", msg)
	}
	return strings.TrimRight(stdout.String(), "\n"), nil
}

func (c CLI) inboxList(ctx context.Context, state string, limit int) (string, error) {
	args := []string{"inbox", "--json"}
	if state != "" {
		args = append(args, "--state", state)
	}
	if limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", limit))
	}
	return c.run(ctx, args...)
}

func (c CLI) jobGet(ctx context.Context, key string) (string, error) {
	return c.run(ctx, "show", key, "--json")
}

func (c CLI) hunkDecide(ctx context.Context, key string, index int, decision, text string) (string, error) {
	args := []string{
		"hunk", "decide", key,
		"--index", fmt.Sprintf("%d", index),
		"--decision", decision,
		"--json",
	}
	if text != "" {
		args = append(args, "--text", text)
	}
	return c.run(ctx, args...)
}

func (c CLI) buildPDF(ctx context.Context, key string) (string, error) {
	return c.run(ctx, "build", key, "--json", "--confirm")
}
