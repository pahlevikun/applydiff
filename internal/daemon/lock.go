package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const lockFileName = "daemon.lock"

// TryLock creates a pid lock in dataDir. Returns release func and held=false when another live daemon holds it.
func TryLock(dataDir string) (release func(), held bool, msg string) {
	path := filepath.Join(dataDir, lockFileName)
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return func() {}, false, ""
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return func() {}, false, ""
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if other, ok := readLockPID(path); ok && processAlive(other) {
			return func() {}, false, fmt.Sprintf("daemon already running (pid %d); exiting", other)
		}
		return func() {}, false, "daemon already running; exiting"
	}
	pid := fmt.Sprintf("%d\n", os.Getpid())
	_ = f.Truncate(0)
	_, _ = f.WriteString(pid)
	release = func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
		_ = os.Remove(path)
	}
	return release, true, ""
}

func readLockPID(path string) (int, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		return 0, false
	}
	return n, true
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil
}
