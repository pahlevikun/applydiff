package webembed

import (
	"embed"
	"io/fs"
	"strings"
)

// Public holds the Nuxt build output copied from web/.output/public by `make web`.
//
//go:embed all:public
var Public embed.FS

const missingMessage = "web assets not built — run `make web`"

// Sub returns the embedded public file system, or nil when assets were not built.
func Sub() (fs.FS, error) {
	sub, err := fs.Sub(Public, "public")
	if err != nil {
		return nil, err
	}
	if !built(sub) {
		return nil, errNotBuilt
	}
	return sub, nil
}

var errNotBuilt = &notBuiltError{}

type notBuiltError struct{}

func (e *notBuiltError) Error() string { return missingMessage }

// MissingMessage is returned when the UI was not embedded at build time.
func MissingMessage() string { return missingMessage }

func built(sub fs.FS) bool {
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return false
	}
	for _, e := range entries {
		name := e.Name()
		if name == ".gitkeep" || name == ".keep" {
			continue
		}
		if name == "index.html" {
			return true
		}
		if e.IsDir() && name == "_nuxt" {
			return true
		}
	}
	// Generated Nuxt output may only expose _nuxt + 200.html, etc.
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".html") && e.Name() != ".gitkeep" {
			return true
		}
	}
	return false
}
