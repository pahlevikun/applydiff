package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type apiRunner struct {
	cfg    Config
	client *http.Client
}

func NewAPIRunner(cfg Config) (Runner, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("llm: LLM_API_KEY required for api provider")
	}
	return &apiRunner{
		cfg: cfg,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

func (a *apiRunner) Complete(ctx context.Context, prompt string) (string, error) {
	body := map[string]interface{}{
		"model": a.cfg.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.cfg.APIURL, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.cfg.APIKey)
	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("llm api: status %d: %s", resp.StatusCode, truncate(string(data), 400))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("llm api: empty choices")
	}
	return parsed.Choices[0].Message.Content, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
