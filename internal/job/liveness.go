package job

import (
	"context"
	"net/http"
	"time"
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Alive probes url with HEAD, then GET if HEAD is not allowed.
func Alive(ctx context.Context, client HTTPDoer, jobURL string) bool {
	if jobURL == "" {
		return false
	}
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	if ok, err := probe(ctx, client, jobURL, http.MethodHead); err == nil && ok {
		return true
	}
	ok, err := probe(ctx, client, jobURL, http.MethodGet)
	return err == nil && ok
}

func probe(ctx context.Context, client HTTPDoer, jobURL, method string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, method, jobURL, nil)
	if err != nil {
		return false, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400, nil
}
