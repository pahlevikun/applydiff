# ApplyDiff

Git-style approval for job applications.

Scan job boards, match each listing against your structured CV, let AI draft the tailoring edits, then review every edit as a diff and approve it hunk by hunk before a LaTeX PDF is built.

## Why

Most tools either auto-rewrite your CV or make you copy-paste. ApplyDiff treats your CV like a codebase: the AI proposes a change, you see exactly what changed and why, and you accept or reject each one.

## Principles

- **Local-first.** Runs on your machine. No account.
- **Human-in-the-loop.** It never submits an application. You click.
- **Structured CV is the source of truth.** LaTeX is an export target.
- **No fabricated numbers.** A metric that is not in your facts file cannot enter a diff.

## Stack

Go + SQLite + Nuxt 3 + LaTeX (latexmk).

## Quick start

```bash
go run ./cmd/applydiff          # API on :8080
cd web && npm run dev           # UI
```

Environment:

| Variable | Purpose |
|----------|---------|
| `DATABASE_PATH` | SQLite file (default `applydiff.db`) |
| `JOBS_CONFIG` | Job sources YAML (default `jobs/config.yaml`) |
| `PORT` | HTTP port (default `8080`) |
| `FACTS_PATH` | JSON facts file (default `facts.json`) |

## LLM (CLI-first)

Tailoring uses a pluggable LLM backend. **No API key is required** if you use a local agent CLI (Claude Code, Codex, etc.) with its own login.

**Selection order**

1. `LLM_PROVIDER` if set (`cli`, `api`, or `fake`)
2. Else `LLM_CLI` if set → `cli`
3. Else auto-detect the first binary on `PATH`: `claude`, `codex`, `cursor-agent`, `hermes`
4. Else `LLM_API_KEY` if set → `api`
5. Else **fake** runner (empty hunks; logs a clear message)

**Supported agent CLIs** (auto-detected defaults):

| Binary | Default command |
|--------|-----------------|
| `claude` | `claude -p` |
| `codex` | `codex exec` |
| `cursor-agent` | `cursor-agent` |
| `hermes` | `hermes` |

Override the command with `LLM_CLI` (stdin receives the prompt; stdout must be the model reply).

**API fallback**

| Variable | Default |
|----------|---------|
| `LLM_API_URL` | `https://api.openai.com/v1/chat/completions` |
| `LLM_MODEL` | `gpt-4o-mini` |
| `LLM_API_KEY` | (required for `api` provider) |

## Database export / import

Move your SQLite data between machines.

**CLI**

```bash
go run ./cmd/applydiff export-db -o backup.db
go run ./cmd/applydiff import-db -i backup.db
```

Uses `VACUUM INTO` for a consistent snapshot. Import replaces the target `DATABASE_PATH`.

**HTTP**

- `GET /api/db/export` — download the SQLite file
- `POST /api/db/import` — request body is the SQLite file (replaces `DATABASE_PATH`; file-backed DB only)

## Migrations

SQL migrations live in `internal/store/migrations/` as `YYYYMMDDHHMMSS_<name>.sql`. They run in lexicographic filename order; applied names are stored in `schema_migrations` and skipped on restart.

## Status

Active development on branch `cursor/applydiff-project-scaffolding-745c`. See `docs/plans/2026-10-14-applydiff.md` and `docs/wireframes/`.
