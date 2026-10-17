# ApplyDiff

Git-style approval for job applications.

Scan job boards, match each listing against your structured CV, let AI draft the tailoring edits, then review every edit as a diff and approve it hunk by hunk before a LaTeX PDF is built.

## Three surfaces (one core)

| Surface | Command |
|---------|---------|
| **CLI** | `applydiff <command>` — scriptable, `--json` / `--dry-run` |
| **TUI** | `applydiff tui` — keyboard-driven review in the terminal |
| **Web** | `applydiff serve` + Nuxt (`cd web && npm run dev`) |

All surfaces call the same Go core in `internal/` and share one SQLite database.

### Quick start

```bash
make cli
./bin/applydiff scan
./bin/applydiff jobs --json
./bin/applydiff tui

./bin/applydiff serve --port 8787   # API on :8787
cd web && npm ci && npm run dev     # UI (proxies /api)
```

Palette: **Phosphor** tokens in `web/assets/tokens-root.css` (mirrored in `internal/theme/theme.css` for TUI).

## Why

Most tools either auto-rewrite your CV or make you copy-paste. ApplyDiff treats your CV like a codebase: the AI proposes a change, you see exactly what changed and why, and you accept or reject each one.

## Principles

- **Local-first.** Runs on your machine. No account.
- **Human-in-the-loop.** It never submits an application. `apply` defaults to **manual**.
- **Structured CV is the source of truth.** LaTeX is an export target.
- **No fabricated numbers.** A metric that is not in your facts file cannot enter a diff.

## Stack

Go + SQLite + Bubble Tea TUI + Nuxt 3 + LaTeX (latexmk).

## CLI commands

`scan`, `jobs`, `show`, `match`, `tailor`, `diff`, `approve`, `reject`, `build`, `apply`, `export`, `import`, `serve`, `tui`

Exit codes: `0` ok · `2` no match · `3` needs approval · `4` LLM error · `5` LaTeX error.

## Environment

| Variable | Purpose |
|----------|---------|
| `DATABASE_PATH` | SQLite file (default `applydiff.db`) |
| `JOBS_CONFIG` | Job sources YAML (default `jobs/config.yaml`) |
| `PORT` | Default HTTP port when running without subcommand (default `8080`) |
| `FACTS_PATH` | JSON facts file (default `facts.json`) |

## LLM (CLI-first)

Tailoring uses a pluggable LLM backend. **No API key is required** if you use a local agent CLI (Claude Code, Codex, etc.) with its own login.

See previous README section for `LLM_PROVIDER`, `LLM_CLI`, and API fallback variables.

## Database export / import

```bash
applydiff export backup.db
applydiff import backup.db
```

Legacy: `export-db` / `import-db` flags still work.

## Migrations

SQL migrations live in `internal/store/migrations/`.

## Status

See `docs/plans/2026-10-17-cli-tui-web.md` and `docs/wireframes/`.
