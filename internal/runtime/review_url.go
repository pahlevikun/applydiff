package runtime

import "fmt"

// ReviewURL returns the web UI diff page for a job key.
func (a *App) ReviewURL(port string, jobKey string) string {
	if port == "" {
		port = "8080"
	}
	if jobKey == "" {
		return fmt.Sprintf("http://127.0.0.1:%s/diff", port)
	}
	return fmt.Sprintf("http://127.0.0.1:%s/diff/%s", port, jobKey)
}
