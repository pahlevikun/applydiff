package source

// Reset clears the registry (tests only).
func Reset() {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = make(map[string]Source)
}

// ByID returns a registered source or nil.
func ByID(id string) Source {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[id]
}
