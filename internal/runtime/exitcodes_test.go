package runtime

import "testing"

func TestExitCodeMapping(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{nil, ExitOK},
		{ErrNoMatch{}, ExitNoMatch},
		{ErrNeedsApproval{Pending: 2}, ExitNeedsApproval},
		{ErrLLM{}, ExitLLM},
		{ErrLaTeX{}, ExitLaTeX},
	}
	for _, c := range cases {
		if got := ExitCode(c.err); got != c.want {
			t.Fatalf("ExitCode(%v)=%d want %d", c.err, got, c.want)
		}
	}
}
