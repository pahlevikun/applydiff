package notify

import (
	"context"
	"strings"
)

// StatusLines reports configured notifiers without leaking secrets.
func StatusLines() []string {
	return []string{
		"stdout: available",
		TelegramStatus(),
	}
}

// FormatStatus is plain text for CLI.
func FormatStatus() string {
	return strings.Join(StatusLines(), "\n")
}

// SendTest delivers a single test message through the primary notifier.
func SendTest(ctx context.Context) error {
	cfg := LoadConfig()
	msg := Message{
		Kind:  "test",
		Title: "ApplyDiff test notification",
		Body:  "If you see this, notifications are wired correctly.",
		Link:  cfg.InboxLink(),
	}
	return PrimaryNotifier().Send(ctx, msg)
}
