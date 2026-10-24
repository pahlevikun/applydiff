package notify

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

// Message is one outbound notification.
type Message struct {
	Kind  string
	Title string
	Body  string
	Link  string
}

// Notifier delivers a message to a channel.
type Notifier interface {
	Send(ctx context.Context, msg Message) error
}

// Config holds notifier wiring from the environment.
type Config struct {
	BaseURL string
}

// LoadConfig reads notification-related env vars.
func LoadConfig() Config {
	base := strings.TrimSpace(os.Getenv("APPLYDIFF_BASE_URL"))
	if base == "" {
		base = "http://localhost:8080"
	}
	return Config{BaseURL: strings.TrimRight(base, "/")}
}

// InboxLink returns the web UI inbox URL.
func (c Config) InboxLink() string {
	return c.BaseURL + "/"
}

// PrimaryNotifier is the notifier used for digest and `notify test` (telegram when configured, else stdout).
func PrimaryNotifier() Notifier {
	if t := NewTelegram(); t.Enabled() {
		return t
	}
	return Stdout{Out: os.Stdout}
}

// DigestNotifier is used by the daemon: telegram only when enabled; otherwise a no-op.
func DigestNotifier() Notifier {
	if t := NewTelegram(); t.Enabled() {
		return t
	}
	return Nop{}
}

// Nop discards messages without error.
type Nop struct{}

func (Nop) Send(context.Context, Message) error { return nil }

// Stdout prints human-readable messages (tests and local dev).
type Stdout struct {
	Out io.Writer
}

func (s Stdout) Send(_ context.Context, msg Message) error {
	w := s.Out
	if w == nil {
		w = os.Stdout
	}
	if msg.Title != "" {
		_, _ = fmt.Fprintf(w, "%s\n", msg.Title)
	}
	if msg.Body != "" {
		_, _ = fmt.Fprintf(w, "%s\n", msg.Body)
	}
	if msg.Link != "" {
		_, _ = fmt.Fprintf(w, "%s\n", msg.Link)
	}
	return nil
}
