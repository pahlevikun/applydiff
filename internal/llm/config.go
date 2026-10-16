package llm

import (
	"os"
	"strings"
)

// Config selects the LLM backend and model.
type Config struct {
	Provider string // "api" or "cli"
	Model    string
	APIURL   string
	APIKey   string
	CLI      string // shell command, e.g. "claude -p"
}

func ConfigFromEnv() Config {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("LLM_PROVIDER")))
	if provider == "" {
		provider = "api"
	}
	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}
	apiURL := os.Getenv("LLM_API_URL")
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1/chat/completions"
	}
	return Config{
		Provider: provider,
		Model:    model,
		APIURL:   apiURL,
		APIKey:   os.Getenv("LLM_API_KEY"),
		CLI:      os.Getenv("LLM_CLI"),
	}
}
