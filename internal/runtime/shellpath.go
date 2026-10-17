package runtime

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

// EnsureLoginShellPATH prepends the login-shell PATH when APPLYDIFF_DESKTOP=1 or when
// started via serve --desktop, so agent CLIs (claude, codex, etc.) remain discoverable
// under a minimal GUI environment.
func EnsureLoginShellPATH() error {
	if os.Getenv("APPLYDIFF_DESKTOP") != "1" {
		return nil
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}
	cmd := exec.Command(shell, "-l", "-c", "printf %s \"$PATH\"")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return err
	}
	loginPath := strings.TrimSpace(out.String())
	if loginPath == "" {
		return nil
	}
	current := os.Getenv("PATH")
	if current == "" {
		os.Setenv("PATH", loginPath)
		return nil
	}
	os.Setenv("PATH", loginPath+string(os.PathListSeparator)+current)
	return nil
}
