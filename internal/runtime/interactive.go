package runtime

import (
	"fmt"
	"os"
)

// RequireInteractiveApproval blocks gated commands unless a TTY, --confirm, or APPLYDIFF_NONINTERACTIVE=1.
func RequireInteractiveApproval(confirm bool) error {
	if confirm {
		return nil
	}
	if os.Getenv("APPLYDIFF_NONINTERACTIVE") == "1" {
		return nil
	}
	fi, err := os.Stdin.Stat()
	if err == nil && (fi.Mode()&os.ModeCharDevice) != 0 {
		return nil
	}
	return fmt.Errorf("approval commands require an interactive terminal (use --confirm or APPLYDIFF_NONINTERACTIVE=1)")
}
