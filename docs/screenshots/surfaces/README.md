# ApplyDiff surface screenshots

Captured **2026-10-17** for the standalone architecture infographic. Paths are relative to this folder.

| Surface | Files | How captured |
|---------|--------|----------------|
| **CLI** | [01-cli.png](01-cli.png) | Styled terminal mock (real `doctor` output shape; async example labeled) |
| **TUI** | [01-tui-jobs.png](01-tui-jobs.png), [01-tui-diff.png](01-tui-diff.png) | Product wireframes (`docs/wireframes/out/`) — run `applydiff tui` locally for live |
| **Web** | [02-web-home.png](02-web-home.png), [02-web-jobs.png](02-web-jobs.png), [02-web-sources.png](02-web-sources.png), [02-web-interview.png](02-web-interview.png) | Headless Chrome vs `applydiff serve` on `:18080` after `make web` |
| **Skill** | [03-skill.png](03-skill.png) | HTML summary of `integrations/skill/SKILL.md` |
| **MCP** | [04-mcp.png](04-mcp.png) | HTML summary of `.cursor-plugin/mcp.json` + tool list |
| **Desktop** | [05-desktop.png](05-desktop.png) | Scaffold status (no Tauri bundle yet) |

## Regenerate web captures

```bash
make web && go build -o bin/applydiff ./cmd/applydiff
./bin/applydiff serve --port 18080
# Chrome headless, e.g.:
"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
  --headless=new --screenshot=docs/screenshots/surfaces/02-web-jobs.png \
  --window-size=1280,800 http://127.0.0.1:18080/jobs
```

HTML mocks for CLI/skill/MCP/desktop live under [preview/](preview/).
