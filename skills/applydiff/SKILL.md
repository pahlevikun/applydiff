---
name: applydiff
description: Review ApplyDiff inbox decisions and CV diff hunks via MCP tools or the CLI. Use when the user wants to see their job queue, inspect fit for a role, record hunk accept/reject decisions, or build a tailored CV PDF. Never submit job applications on the user's behalf.
---

# ApplyDiff (agent on-ramp)

ApplyDiff is a local job-hunting cockpit: a background daemon scans boards, scores roles against a master CV, and drafts small CV edits as hunks you review. Humans (or you, on their instruction) decide each hunk; nothing is sent to employers automatically.

Product docs and setup: [github.com/pahlevikun/applydiff](https://github.com/pahlevikun/applydiff)

## MCP tools (preferred)

Connect `applydiff mcp` (stdio). Each tool returns the same JSON as the matching `applydiff … --json` command—no extra fields.

| Tool | When to use |
|------|-------------|
| `inbox_list` | See the decision queue (`state?`, `limit?`). Start here. |
| `job_get` | Open one job: fit score, dimensions, requirements, risk. |
| `hunk_decide` | Record `accept`, `reject`, `edit`, or `skip` for hunk `index` on `key`. |
| `build_pdf` | After every hunk is decided, compile the PDF (`path`) or read `blocked` + `undecided`. |

## Rules

- **Never submit an application.** Do not run `applydiff apply` or use any submit/auto-send path. If the user wants to mark something sent, give them the exact command to run locally.
- **Never auto-approve hunks.** Only call `hunk_decide` when the user explicitly chooses accept/reject/edit/skip for that hunk.
- **No business logic in the skill.** Delegate to MCP tools or CLI `--json`; do not re-score, re-rank, or invent defaults.

## CLI fallback

```bash
applydiff inbox --json
applydiff show <key> --json
applydiff hunk decide <key> --index N --decision accept|reject|edit|skip [--text ...] --json
applydiff build <key> --json
```

## Empty inbox

If `inbox_list` is empty and counts are zero, the daemon may not be running. Ask the user to start it (or run locally):

```bash
applydiff daemon
```

Then wait for a scan cycle or run `applydiff scan --json` if they want fresh jobs immediately.
