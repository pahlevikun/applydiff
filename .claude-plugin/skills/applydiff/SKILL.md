---
name: applydiff
description: Route job search, tailoring, and human approval for ApplyDiff CLI workflows.
---

# ApplyDiff router

Use the `applydiff` CLI (JSON via `--json`). Never run `approve`, `reject`, `build`, or `apply` — those require a human in the terminal or web UI.

## Modes

| Mode | Commands |
|------|----------|
| setup | `applydiff doctor --json`, `applydiff init` |
| scan | `applydiff scan --json` |
| jobs | `applydiff jobs --json`, `applydiff show <key> --json` |
| match | `applydiff match <key> --json` |
| tailor | `applydiff add` (stdin JD) → `match` → `tailor <key> --context --json` → host drafts → `propose <key> --hunks -` |
| review | `applydiff open <key>` — hand off to human |
| build | Tell user to run `applydiff build` locally after approving hunks |
| interview | `applydiff interview kit <key> --json` |
| doctor | `applydiff doctor --json` |

## Pasted job description

1. `applydiff add --title "..." --company "..." --url "..."` with JD on stdin
2. `applydiff match <id> --json`
3. `applydiff tailor <id> --context --json` then `applydiff propose <id> --hunks -`
4. `applydiff open <id>` for human review
