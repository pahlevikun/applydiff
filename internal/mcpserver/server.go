package mcpserver

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

const protocolVersion = "2024-11-05"

// Run serves MCP over stdio (JSON-RPC 2.0, newline-delimited). Only JSON-RPC frames go to stdout.
func Run(ctx context.Context) error {
	fmt.Fprintln(os.Stderr, "applydiff mcp: ready")
	srv := &Server{cli: CLI{Bin: ResolveCLI()}, out: os.Stdout, encMu: new(sync.Mutex)}
	return srv.serve(ctx, os.Stdin)
}

// Server implements a minimal MCP stdio server by delegating to the CLI.
type Server struct {
	cli   CLI
	out   io.Writer
	encMu *sync.Mutex
}

func (s *Server) serve(ctx context.Context, in io.Reader) error {
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		if err := s.handleLine(ctx, sc.Bytes()); err != nil {
			return err
		}
	}
	return sc.Err()
}

func (s *Server) handleLine(ctx context.Context, line []byte) error {
	line = trimSpace(line)
	if len(line) == 0 {
		return nil
	}
	var req rpcRequest
	if err := json.Unmarshal(line, &req); err != nil {
		return s.writeResponse(rpcResponse{JSONRPC: "2.0", ID: nil, Error: &rpcError{Code: -32700, Message: "parse error"}})
	}
	if req.JSONRPC != "2.0" {
		return s.writeResponse(rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: -32600, Message: "invalid request"}})
	}
	if req.ID == nil {
		return s.handleNotification(ctx, req.Method, req.Params)
	}
	result, errObj := s.dispatch(ctx, req.Method, req.Params)
	if errObj != nil {
		return s.writeResponse(rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: errObj})
	}
	return s.writeResponse(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: result})
}

func (s *Server) handleNotification(ctx context.Context, method string, params json.RawMessage) error {
	switch method {
	case "notifications/initialized", "initialized":
		return nil
	default:
		return nil
	}
}

func (s *Server) dispatch(ctx context.Context, method string, params json.RawMessage) (any, *rpcError) {
	switch method {
	case "initialize":
		return s.handleInitialize(params)
	case "tools/list":
		return map[string]any{"tools": toolDefinitions()}, nil
	case "tools/call":
		return s.handleToolsCall(ctx, params)
	default:
		return nil, &rpcError{Code: -32601, Message: "method not found"}
	}
}

func (s *Server) handleInitialize(params json.RawMessage) (any, *rpcError) {
	return map[string]any{
		"protocolVersion": protocolVersion,
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]string{
			"name":    "applydiff",
			"version": "dev",
		},
	}, nil
}

type toolsCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (s *Server) handleToolsCall(ctx context.Context, params json.RawMessage) (any, *rpcError) {
	var p toolsCallParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcError{Code: -32602, Message: "invalid params"}
	}
	text, isErr, err := s.callTool(ctx, p.Name, p.Arguments)
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
		"isError": isErr,
	}, nil
}

func (s *Server) callTool(ctx context.Context, name string, rawArgs json.RawMessage) (text string, isErr bool, err error) {
	switch name {
	case "inbox_list":
		var args struct {
			State string `json:"state"`
			Limit int    `json:"limit"`
		}
		if len(rawArgs) > 0 && string(rawArgs) != "null" {
			if err := json.Unmarshal(rawArgs, &args); err != nil {
				return "", false, fmt.Errorf("invalid arguments: %w", err)
			}
		}
		out, err := s.cli.inboxList(ctx, args.State, args.Limit)
		return out, false, err
	case "job_get":
		var args struct {
			Key string `json:"key"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil || args.Key == "" {
			return "", false, fmt.Errorf("key required")
		}
		out, err := s.cli.jobGet(ctx, args.Key)
		return out, false, err
	case "hunk_decide":
		var args struct {
			Key      string `json:"key"`
			Index    int    `json:"index"`
			Decision string `json:"decision"`
			Text     string `json:"text"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return "", false, fmt.Errorf("invalid arguments: %w", err)
		}
		if args.Key == "" || args.Index < 1 {
			return "", false, fmt.Errorf("key and index required")
		}
		switch args.Decision {
		case "accept", "reject", "edit", "skip":
		default:
			return "", false, fmt.Errorf("unknown decision %q", args.Decision)
		}
		out, err := s.cli.hunkDecide(ctx, args.Key, args.Index, args.Decision, args.Text)
		return out, false, err
	case "build_pdf":
		var args struct {
			Key string `json:"key"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil || args.Key == "" {
			return "", false, fmt.Errorf("key required")
		}
		out, err := s.cli.buildPDF(ctx, args.Key)
		return out, false, err
	default:
		return "", false, fmt.Errorf("unknown tool %q", name)
	}
}

func toolDefinitions() []map[string]any {
	return []map[string]any{
		{
			"name":        "inbox_list",
			"description": "List the decision queue (same JSON as applydiff inbox --json)",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"state": map[string]any{"type": "string", "description": "needs-decision, in-progress, sent, interview"},
					"limit": map[string]any{"type": "integer", "minimum": 0},
				},
			},
		},
		{
			"name":        "job_get",
			"description": "Job detail with match fit dimensions (applydiff show --json)",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key": map[string]any{"type": "string"},
				},
				"required": []string{"key"},
			},
		},
		{
			"name":        "hunk_decide",
			"description": "Record accept/reject/edit/skip for one hunk (applydiff hunk decide --json)",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key":      map[string]any{"type": "string"},
					"index":    map[string]any{"type": "integer", "minimum": 1},
					"decision": map[string]any{"type": "string", "enum": []string{"accept", "reject", "edit", "skip"}},
					"text":     map[string]any{"type": "string"},
				},
				"required": []string{"key", "index", "decision"},
			},
		},
		{
			"name":        "build_pdf",
			"description": "Build CV PDF when all hunks are decided (applydiff build --json)",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key": map[string]any{"type": "string"},
				},
				"required": []string{"key"},
			},
		},
	}
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s *Server) writeResponse(resp rpcResponse) error {
	s.encMu.Lock()
	defer s.encMu.Unlock()
	enc := json.NewEncoder(s.out)
	return enc.Encode(resp)
}

func trimSpace(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t' || b[0] == '\r' || b[0] == '\n') {
		b = b[1:]
	}
	for len(b) > 0 && (b[len(b)-1] == ' ' || b[len(b)-1] == '\t' || b[len(b)-1] == '\r' || b[len(b)-1] == '\n') {
		b = b[:len(b)-1]
	}
	return b
}

// InstallMCP writes client MCP config snippets (cursor, claude, codex).
func InstallMCP(client string, bin string) error {
	if bin == "" {
		bin = "applydiff"
	}
	entry := map[string]any{
		"command": bin,
		"args":    []string{"mcp"},
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	switch client {
	case "cursor":
		return mergeMCPFile(home+"/.cursor/mcp.json", "applydiff", entry)
	case "claude":
		return mergeMCPFile(home+"/.claude/.mcp.json", "applydiff", entry)
	case "codex":
		return mergeMCPFile(home+"/.codex/mcp.json", "applydiff", entry)
	default:
		return fmt.Errorf("unknown mcp client %q", client)
	}
}

func mergeMCPFile(path, name string, entry map[string]any) error {
	var cfg map[string]any
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &cfg)
	}
	if cfg == nil {
		cfg = map[string]any{}
	}
	servers, _ := cfg["mcpServers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	servers[name] = entry
	cfg["mcpServers"] = servers
	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}
