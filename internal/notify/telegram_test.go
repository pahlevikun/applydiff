package notify_test

import (
	"strings"
	"testing"

	"github.com/pahlevikun/applydiff/internal/notify"
)

func TestTelegramStatusNoToken(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	got := notify.TelegramStatus()
	if got != "telegram: disabled (no token)" {
		t.Fatalf("got %q", got)
	}
}

func TestTelegramStatusMasksChatID(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "secret-token")
	t.Setenv("TELEGRAM_CHAT_ID", "123456789")
	got := notify.TelegramStatus()
	if strings.Contains(got, "secret") {
		t.Fatalf("token leaked in status: %q", got)
	}
	if !strings.Contains(got, "enabled") {
		t.Fatalf("expected enabled: %q", got)
	}
}
