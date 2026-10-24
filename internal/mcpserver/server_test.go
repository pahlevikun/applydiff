package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestMCPInitializeAndToolsList(t *testing.T) {
	cli := writeFakeCLI(t)
	t.Setenv("APPLYDIFF_BIN", cli)

	var stdout bytes.Buffer
	srv := &Server{cli: CLI{Bin: cli}, out: &stdout, encMu: new(sync.Mutex)}
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
`)
	if err := srv.serve(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 JSON-RPC lines on stdout, got %d:\n%s", len(lines), stdout.String())
	}
	for i, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Fatalf("line %d not valid JSON: %s", i, line)
		}
	}
	var listResp struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[1]), &listResp); err != nil {
		t.Fatal(err)
	}
	if len(listResp.Result.Tools) != 4 {
		t.Fatalf("expected 4 tools, got %d", len(listResp.Result.Tools))
	}
	names := map[string]bool{}
	for _, tool := range listResp.Result.Tools {
		names[tool.Name] = true
	}
	for _, want := range []string{"inbox_list", "job_get", "hunk_decide", "build_pdf"} {
		if !names[want] {
			t.Fatalf("missing tool %s in %v", want, names)
		}
	}
}

func TestMCPStdoutOnlyJSONRPC(t *testing.T) {
	cli := writeFakeCLI(t)
	t.Setenv("APPLYDIFF_BIN", cli)

	oldStderr := os.Stderr
	_, r, _ := os.Pipe()
	os.Stderr = r
	defer func() { os.Stderr = oldStderr }()

	var stdout bytes.Buffer
	srv := &Server{cli: CLI{Bin: cli}, out: &stdout, encMu: new(sync.Mutex)}
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}
`)
	if err := srv.serve(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	if strings.Contains(out, "ready") {
		t.Fatalf("stdout must not contain readiness log: %q", out)
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		var probe map[string]any
		if err := json.Unmarshal([]byte(line), &probe); err != nil {
			t.Fatalf("stdout line is not JSON-RPC: %q err=%v", line, err)
		}
		if probe["jsonrpc"] != "2.0" {
			t.Fatalf("missing jsonrpc field: %q", line)
		}
	}
}

func TestMCPToolCallsUseCLIPayload(t *testing.T) {
	cli := writeFakeCLI(t)
	srv := &Server{cli: CLI{Bin: cli}, out: ioDiscard(), encMu: new(sync.Mutex)}

	cases := []struct {
		tool string
		args string
		want string
	}{
		{"inbox_list", `{}`, `{"data":[],"meta":{"counts":{}}}`},
		{"job_get", `{"key":"job-1"}`, `{"job":{"id":"job-1"}}`},
		{"hunk_decide", `{"key":"job-1","index":1,"decision":"accept"}`, `{"decided":1,"total":2}`},
		{"build_pdf", `{"key":"job-1"}`, `{"cv_pdf":"/tmp/out.pdf"}`},
	}
	for _, tc := range cases {
		text, isErr, err := srv.callTool(context.Background(), tc.tool, []byte(tc.args))
		if err != nil || isErr {
			t.Fatalf("%s: err=%v isErr=%v", tc.tool, err, isErr)
		}
		if text != tc.want {
			t.Fatalf("%s: got %q want %q", tc.tool, text, tc.want)
		}
	}
}

func TestMCPHunkDecideRejectsUnknownDecision(t *testing.T) {
	cli := writeFakeCLI(t)
	srv := &Server{cli: CLI{Bin: cli}, out: ioDiscard(), encMu: new(sync.Mutex)}
	_, _, err := srv.callTool(context.Background(), "hunk_decide", []byte(`{"key":"job-1","index":1,"decision":"maybe"}`))
	if err == nil || !strings.Contains(err.Error(), "unknown decision") {
		t.Fatalf("expected unknown decision error, got %v", err)
	}
	_, _, err = srv.callTool(context.Background(), "hunk_decide", []byte(`{"key":"job-1","index":1,"decision":"reject"}`))
	if err != nil {
		t.Fatalf("reject: %v", err)
	}
}

func writeFakeCLI(t *testing.T) string {
	dir := t.TempDir()
	path := filepath.Join(dir, "applydiff-fake")
	script := `#!/bin/sh
set -e
case "$1" in
inbox)
  echo '{"data":[],"meta":{"counts":{}}}'
  ;;
show)
  echo '{"job":{"id":"job-1"}}'
  ;;
hunk)
  if [ "$2" != decide ]; then exit 2; fi
  key="$3"
  shift 3
  decision=""
  while [ $# -gt 0 ]; do
    case "$1" in
    --index) shift; shift ;;
    --decision) decision="$2"; shift 2 ;;
    --text) shift 2 ;;
    --json) shift ;;
    *) shift ;;
    esac
  done
  case "$decision" in
  accept) echo '{"decided":1,"total":2}' ;;
  reject) echo '{"decided":1,"total":2}' ;;
  *) echo "invalid decision" >&2; exit 1 ;;
  esac
  ;;
build)
  echo '{"cv_pdf":"/tmp/out.pdf"}'
  ;;
*)
  echo "unknown command $1" >&2
  exit 1
  ;;
esac
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func ioDiscard() *bytes.Buffer {
	return &bytes.Buffer{}
}
