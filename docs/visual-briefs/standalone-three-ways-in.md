# Visual brief: Standalone three ways in

## Evidence pack

| Claim | Status |
|-------|--------|
| Five surfaces (CLI/TUI, web, skill, MCP, desktop) | verified — plan + `cmd/applydiff`, `internal/mcpserver`, `integrations/` |
| Single `runtime.App` + SQLite in app-data | verified — `internal/runtime`, `internal/paths` |
| Agent drafts via `--context` + `propose`; human approves in UI/TTY | verified — skill, `gate.go`, `tailor_propose.go` |
| Tauri desktop | inference — scaffold only (`desktop/README.md`) |
| Install via curl/brew templates + GoReleaser | verified — `scripts/install.sh`, `.goreleaser.yaml` |
| Backlog items (signing, tectonic bin download, version stamp) | verified — audit notes in plan follow-up |

## Takeaway

> After five seconds, the viewer should understand that every entry point hits the same Go core and database, but only a human can approve CV hunks and build the PDF.

## Visual job + layout

- **Job:** Map relationships and flow (ways in → core → human gate).
- **Family:** Convergence (many inputs → one core) plus a **binary-comparison** strip (agent path vs human path).

## Platform

- **Dimensions:** 1200×720 SVG (README / docs / slide).
- **Production:** Deterministic SVG, Phosphor editorial tokens aligned with `assets/editorial-pipeline-convergence.svg`.

## Overlay copy (on asset)

- Headline: *Five ways in funnel into one Go core and one SQLite file on your machine*
- Sub: *Agents may draft hunks via skill or MCP; approve, reject, build, and apply stay human-gated.*

## Asset path

`assets/editorial-standalone-three-ways-in.svg`

## QA (anti-slop)

| Check | Score |
|-------|-------|
| Clear visual job | pass |
| Evidence labeled | pass (footer + brief) |
| Not title-only | pass |
| Distinct from pipeline convergence asset | pass (install + MCP, not scan funnel) |
| Fake UI | none |

## Limitations

- Desktop shown as dashed scaffold, not a product screenshot.
- Does not enumerate every MCP tool name (see `internal/mcpserver/server.go`).

## Surface screenshots (batch)

Per-surface PNGs: [docs/screenshots/surfaces/README.md](../screenshots/surfaces/README.md).
