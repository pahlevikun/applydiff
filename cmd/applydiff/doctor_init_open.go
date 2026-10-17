package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	goruntime "runtime"
	"strings"
	"time"

	"github.com/pahlevikun/applydiff/internal/runtime"
	"github.com/spf13/cobra"
)

func cmdDoctor() *cobra.Command {
	c := &cobra.Command{
		Use:   "doctor",
		Short: "Check data folder, database, LLM, and PDF toolchain",
		Run: func(cmd *cobra.Command, args []string) {
			rep, err := runtime.Doctor(cmd.Context(), rootApp)
			exitOn(err)
			exitOn(emitOrJSON(rep, func() string {
				var b strings.Builder
				b.WriteString(fmt.Sprintf("home: %s\n", rep.Home))
				b.WriteString(fmt.Sprintf("db: %s (ok=%v)\n", rep.DB, rep.DBOK))
				b.WriteString(fmt.Sprintf("config: %s\n", rep.Config))
				b.WriteString(fmt.Sprintf("llm: %s", rep.LLMProvider))
				if rep.LLMCLI != "" {
					b.WriteString(fmt.Sprintf(" cli=%s", rep.LLMCLI))
				}
				b.WriteString("\n")
				b.WriteString(fmt.Sprintf("pdf: %s\n", rep.PDFEngine))
				for k, v := range rep.MCP {
					b.WriteString(fmt.Sprintf("mcp %s: %s\n", k, v))
				}
				if rep.DesktopPATH {
					b.WriteString("desktop: login-shell PATH enabled (APPLYDIFF_DESKTOP=1)\n")
				}
				return strings.TrimRight(b.String(), "\n")
			}))
		},
	}
	return c
}

func cmdInit() *cobra.Command {
	c := &cobra.Command{
		Use:   "init",
		Short: "Create the data folder and default jobs config",
		Run: func(cmd *cobra.Command, args []string) {
			p, err := runtime.InitData()
			exitOn(err)
			if rootApp != nil {
				p = rootApp.Paths
			}
			exitOn(emitOrJSON(map[string]string{
				"home":   p.Home,
				"db":     p.DB,
				"config": p.Config,
			}, func() string {
				return fmt.Sprintf("initialized %s", p.Home)
			}))
		},
	}
	return c
}

func cmdOpen() *cobra.Command {
	var port string
	c := &cobra.Command{
		Use:   "open [job-key]",
		Short: "Open the web UI (and job diff page when a key is given)",
		Run: func(cmd *cobra.Command, args []string) {
			if port == "" {
				port = os.Getenv("PORT")
				if port == "" {
					port = "8080"
				}
			}
			base := fmt.Sprintf("http://127.0.0.1:%s", port)
			path := "/"
			if len(args) > 0 {
				_, err := rootApp.ShowJob(cmd.Context(), args[0])
				exitOn(err)
				path = "/diff/" + args[0]
			}
			child, started, err := ensureServeRunning(port)
			exitOn(err)
			url := base + path
			exitOn(openBrowser(url))
			if started && child != nil {
				fmt.Fprintf(os.Stderr, "started serve on %s (pid %d); leave it running or stop manually\n", base, child.Process.Pid)
			} else {
				fmt.Fprintf(os.Stderr, "opened %s\n", url)
			}
		},
	}
	c.Flags().StringVar(&port, "port", "", "serve port (default 8080 or PORT)")
	return c
}

func ensureServeRunning(port string) (*exec.Cmd, bool, error) {
	if serverHealthy(port) {
		return nil, false, nil
	}
	exe, err := os.Executable()
	if err != nil {
		return nil, false, err
	}
	child := exec.Command(exe, "serve", "--port", port)
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr
	if err := child.Start(); err != nil {
		return nil, false, err
	}
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if serverHealthy(port) {
			return child, true, nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	child.Process.Kill()
	return nil, false, fmt.Errorf("serve did not become healthy on port %s", port)
}

func serverHealthy(port string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%s/healthz", port), nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
