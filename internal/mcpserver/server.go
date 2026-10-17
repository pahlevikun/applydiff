package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pahlevikun/applydiff/internal/features"
	"github.com/pahlevikun/applydiff/internal/runtime"
	"github.com/pahlevikun/applydiff/internal/tailor"
)

// Run serves MCP tools over stdio using the shared runtime.App.
func Run(ctx context.Context, app *runtime.App) error {
	server := mcp.NewServer(&mcp.Implementation{Name: "applydiff", Version: "dev"}, nil)

	type jobKeyArgs struct {
		JobKey string `json:"job_key" jsonschema:"job id or key"`
	}
	type empty struct{}

	mcp.AddTool(server, &mcp.Tool{Name: "list_jobs", Description: "List jobs with filter scores"}, func(ctx context.Context, _ *mcp.CallToolRequest, _ empty) (*mcp.CallToolResult, any, error) {
		jobs, err := app.ListJobs(ctx)
		return jsonResult(jobs, err)
	})

	mcp.AddTool(server, &mcp.Tool{Name: "show_job", Description: "Job detail and match"}, func(ctx context.Context, _ *mcp.CallToolRequest, args jobKeyArgs) (*mcp.CallToolResult, any, error) {
		d, err := app.ShowJob(ctx, args.JobKey)
		return jsonResult(d, err)
	})

	type addJobArgs struct {
		Title       string `json:"title"`
		Company     string `json:"company"`
		URL         string `json:"url"`
		Description string `json:"description"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "add_job", Description: "Add pasted job description"}, func(ctx context.Context, _ *mcp.CallToolRequest, args addJobArgs) (*mcp.CallToolResult, any, error) {
		os.Setenv("APPLYDIFF_ACTOR", "agent:mcp")
		j, err := app.AddJob(ctx, runtime.AddJobInput{
			Title: args.Title, Company: args.Company, URL: args.URL, Description: args.Description,
		}, runtime.Actor("agent:mcp"))
		return jsonResult(j, err)
	})

	mcp.AddTool(server, &mcp.Tool{Name: "scan", Description: "Scan configured job sources (async)"}, func(ctx context.Context, _ *mcp.CallToolRequest, _ empty) (*mcp.CallToolResult, any, error) {
		run, err := app.Runs.Start(ctx, "scan", "", func(ctx context.Context) (any, error) {
			return app.Scan(ctx, runtime.ScanOptions{})
		})
		return jsonResult(run, err)
	})

	mcp.AddTool(server, &mcp.Tool{Name: "match", Description: "Compute match for a job"}, func(ctx context.Context, _ *mcp.CallToolRequest, args jobKeyArgs) (*mcp.CallToolResult, any, error) {
		run, err := app.Runs.Start(ctx, "match", args.JobKey, func(ctx context.Context) (any, error) {
			return app.ComputeMatch(ctx, args.JobKey, false)
		})
		return jsonResult(run, err)
	})

	mcp.AddTool(server, &mcp.Tool{Name: "tailor", Description: "Tailor with built-in LLM (async)"}, func(ctx context.Context, _ *mcp.CallToolRequest, args jobKeyArgs) (*mcp.CallToolResult, any, error) {
		run, err := app.Runs.Start(ctx, "tailor", args.JobKey, func(ctx context.Context) (any, error) {
			return app.Tailor(ctx, args.JobKey, false)
		})
		return jsonResult(run, err)
	})

	mcp.AddTool(server, &mcp.Tool{Name: "tailor_context", Description: "Context for host-agent drafting"}, func(ctx context.Context, _ *mcp.CallToolRequest, args jobKeyArgs) (*mcp.CallToolResult, any, error) {
		tc, err := app.TailorContext(ctx, args.JobKey)
		return jsonResult(tc, err)
	})

	type proposeArgs struct {
		JobKey string           `json:"job_key"`
		Hunks  []tailor.RawHunk `json:"hunks"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "propose_hunks", Description: "Save validated hunks from host agent"}, func(ctx context.Context, _ *mcp.CallToolRequest, args proposeArgs) (*mcp.CallToolResult, any, error) {
		prop, err := app.ProposeHunks(ctx, args.JobKey, args.Hunks, runtime.Actor("agent:mcp"))
		return jsonResult(prop, err)
	})

	type runIDArgs struct {
		RunID string `json:"run_id"`
	}
	mcp.AddTool(server, &mcp.Tool{Name: "get_run", Description: "Async run status"}, func(ctx context.Context, _ *mcp.CallToolRequest, args runIDArgs) (*mcp.CallToolResult, any, error) {
		run, err := app.Runs.Get(ctx, args.RunID)
		return jsonResult(run, err)
	})

	mcp.AddTool(server, &mcp.Tool{Name: "review_link", Description: "Web UI diff URL"}, func(ctx context.Context, _ *mcp.CallToolRequest, args jobKeyArgs) (*mcp.CallToolResult, any, error) {
		url := app.ReviewURL("8080", args.JobKey)
		return jsonResult(map[string]string{"url": url}, nil)
	})

	if features.InterviewPrepEnabled {
		mcp.AddTool(server, &mcp.Tool{Name: "interview_kit", Description: "Interview prep surface for job"}, func(ctx context.Context, _ *mcp.CallToolRequest, args jobKeyArgs) (*mcp.CallToolResult, any, error) {
			kit, err := app.GenerateInterviewKit(ctx, args.JobKey)
			return jsonResult(kit, err)
		})
	}

	mcp.AddTool(server, &mcp.Tool{Name: "doctor", Description: "Environment diagnostics"}, func(ctx context.Context, _ *mcp.CallToolRequest, _ empty) (*mcp.CallToolResult, any, error) {
		rep, err := runtime.Doctor(ctx, app)
		return jsonResult(rep, err)
	})

	return server.Run(ctx, &mcp.StdioTransport{})
}

func jsonResult(v any, err error) (*mcp.CallToolResult, any, error) {
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}, IsError: true}, nil, nil
	}
	raw, _ := json.MarshalIndent(v, "", "  ")
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(raw)}}}, v, nil
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
