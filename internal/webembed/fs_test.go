package webembed

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestBuiltRequiresRealAssets(t *testing.T) {
	empty := fstest.MapFS{
		".gitkeep": &fstest.MapFile{Data: []byte("")},
	}
	if built(empty) {
		t.Fatal("expected placeholder-only dir to be not built")
	}
	withIndex := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<!DOCTYPE html><html></html>")},
	}
	if !built(withIndex) {
		t.Fatal("expected index.html to count as built")
	}
}

func TestSubWhenBuilt(t *testing.T) {
	sub, err := Sub()
	if err != nil {
		if _, ok := err.(*notBuiltError); ok {
			t.Skip("web assets not embedded in this build")
		}
		t.Fatal(err)
	}
	_, err = fs.ReadFile(sub, "index.html")
	if err != nil {
		t.Fatalf("index.html: %v", err)
	}
}
