package llm

import "context"

// FakeRunner returns a fixed response for tests.
type FakeRunner struct {
	Response string
	Err      error
}

func (f FakeRunner) Complete(ctx context.Context, prompt string) (string, error) {
	if f.Err != nil {
		return "", f.Err
	}
	return f.Response, nil
}
