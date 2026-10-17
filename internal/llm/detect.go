package llm

import (
	"os/exec"
	"strings"
)

type knownCLI struct {
	binary  string
	command string
}

var knownCLIs = []knownCLI{
	{binary: "claude", command: "claude -p"},
	{binary: "codex", command: "codex exec"},
	{binary: "cursor-agent", command: "cursor-agent"},
	{binary: "hermes", command: "hermes"},
}

// DetectCLI returns a default shell command for the first known agent binary on PATH.
func DetectCLI() string {
	for _, k := range knownCLIs {
		if path, err := exec.LookPath(k.binary); err == nil && strings.TrimSpace(path) != "" {
			return k.command
		}
	}
	return ""
}
