package runtime

// Exit codes for CLI scripting.
const (
	ExitOK            = 0
	ExitNoMatch       = 2
	ExitNeedsApproval = 3
	ExitLLM           = 4
	ExitLaTeX         = 5
)

// ExitCode maps errors to process exit codes.
func ExitCode(err error) int {
	if err == nil {
		return ExitOK
	}
	switch {
	case isErrNoMatch(err):
		return ExitNoMatch
	case isErrNeedsApproval(err):
		return ExitNeedsApproval
	case isErrLLM(err):
		return ExitLLM
	case isErrLaTeX(err):
		return ExitLaTeX
	default:
		return 1
	}
}

func isErrNoMatch(err error) bool {
	_, ok := err.(ErrNoMatch)
	return ok
}

func isErrNeedsApproval(err error) bool {
	_, ok := err.(ErrNeedsApproval)
	return ok
}

func isErrLLM(err error) bool {
	_, ok := err.(ErrLLM)
	return ok
}

func isErrLaTeX(err error) bool {
	_, ok := err.(ErrLaTeX)
	return ok
}
