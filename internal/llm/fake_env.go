package llm

import (
	"os"
	"strings"
)

const envFakeResponse = "LLM_FAKE_RESPONSE"

func fakeResponseFromEnv() string {
	if v := strings.TrimSpace(os.Getenv(envFakeResponse)); v != "" {
		return v
	}
	return `{"hunks":[]}`
}
