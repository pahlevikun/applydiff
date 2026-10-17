package llm

import (
	"os"
	"strings"
)

// Config selects the LLM backend and model.
type Config struct {
	Provider string // "api", "cli", or "fake"
	Model    string
	APIURL   string
	APIKey   string
	CLI      string // shell command, e.g. "claude -p"
}

func ConfigFromEnv() Config {
	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}
	apiURL := os.Getenv("LLM_API_URL")
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1/chat/completions"
	}
	cfg := Config{
		Model:  model,
		APIURL: apiURL,
		APIKey: strings.TrimSpace(os.Getenv("LLM_API_KEY")),
		CLI:    strings.TrimSpace(os.Getenv("LLM_CLI")),
	}

	explicit := strings.ToLower(strings.TrimSpace(os.Getenv("LLM_PROVIDER")))
	if explicit != "" {
		cfg.Provider = explicit
		if cfg.Provider == "cli" && cfg.CLI == "" {
			cfg.CLI = DetectCLI()
		}
		return cfg
	}
	if cfg.CLI != "" {
		cfg.Provider = "cli"
		return cfg
	}
	if cmd := DetectCLI(); cmd != "" {
		cfg.CLI = cmd
		cfg.Provider = "cli"
		return cfg
	}
	if cfg.APIKey != "" {
		cfg.Provider = "api"
		return cfg
	}
	cfg.Provider = "fake"
	return cfg
}
