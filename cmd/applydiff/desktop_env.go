package main

import (
	"os"
)

// prepareDesktopEnv sets APPLYDIFF_DESKTOP when serve is invoked with --desktop so
// runtime.EnsureLoginShellPATH runs before LLM CLI detection.
func prepareDesktopEnv() {
	if !argsContain("--desktop") {
		return
	}
	_ = os.Setenv("APPLYDIFF_DESKTOP", "1")
}

func argsContain(flag string) bool {
	for _, a := range os.Args[1:] {
		if a == flag {
			return true
		}
	}
	return false
}
