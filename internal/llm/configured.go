package llm

// BackendConfigured reports whether a real LLM backend is configured (not degraded fake).
func BackendConfigured() bool {
	cfg := ConfigFromEnv()
	return cfg.Provider != "fake"
}
