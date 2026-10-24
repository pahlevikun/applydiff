package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Telegram sends messages via the Bot API.
type Telegram struct {
	Token  string
	ChatID string
	Client *http.Client
}

// NewTelegram reads TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID.
func NewTelegram() Telegram {
	return Telegram{
		Token:  strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
		ChatID: strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID")),
		Client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Enabled is true when both token and chat id are set.
func (t Telegram) Enabled() bool {
	return t.Token != "" && t.ChatID != ""
}

func (t Telegram) Send(ctx context.Context, msg Message) error {
	if !t.Enabled() {
		return nil
	}
	text := formatTelegramText(msg)
	body, err := json.Marshal(map[string]string{
		"chat_id": t.ChatID,
		"text":    text,
	})
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.Token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := t.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram: HTTP %d", resp.StatusCode)
	}
	return nil
}

func formatTelegramText(msg Message) string {
	var b strings.Builder
	if msg.Title != "" {
		b.WriteString(msg.Title)
	}
	if msg.Body != "" {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(msg.Body)
	}
	if msg.Link != "" {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(msg.Link)
	}
	return b.String()
}

// TelegramStatus describes telegram configuration without leaking the token.
func TelegramStatus() string {
	t := NewTelegram()
	if t.Token == "" {
		return "telegram: disabled (no token)"
	}
	if t.ChatID == "" {
		return "telegram: disabled (no chat id)"
	}
	return "telegram: enabled (chat " + maskChatID(t.ChatID) + ")"
}

func maskChatID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) <= 4 {
		return "***"
	}
	return "***" + id[len(id)-4:]
}
