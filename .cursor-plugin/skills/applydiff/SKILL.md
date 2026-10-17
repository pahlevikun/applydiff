---
name: applydiff
description: Search jobs, score them against a master CV, and draft small CV edits with the ApplyDiff CLI or MCP tools. Use when the user asks to scan job boards, paste a job description, match a posting, tailor a CV or resume, or open a job diff. A person must approve, reject, build the PDF, and record the application. Do not run those commands.
---

# ApplyDiff

Use the ApplyDiff MCP tools when they are connected. If they are not, run the `applydiff` CLI and add `--json`.

A person reviews every edit. Do not run `approve`, `reject`, `build`, or `apply`. A hook blocks those commands. Tell the user the exact command to run in their own terminal or in the web UI.

## Start

1. If `applydiff` is not on `PATH`, stop and tell the user to install it.
2. Run `applydiff doctor --json` (MCP: `doctor`).
3. If the data folder is missing, run `applydiff init`.

## What the user asked

| They want | Do this |
|---|---|
| Check setup | `doctor`, then `init` if needed |
| Pull jobs from boards | `applydiff scan --json`, then `applydiff jobs --json` |
| See one job | `applydiff show <key> --json` |
| Score a job against the CV | `applydiff match <key> --json` |
| Paste a job description | `add`, then `match`, then draft edits |
| Draft CV edits | `tailor`, or `tailor --context` plus `propose` |
| Review the diff | `applydiff open <key>` (MCP: `review_link`) |
| Build the PDF or mark applied | Tell them the command. Do not run it. |

`scan`, `match`, and `tailor` can take a while. MCP runs `scan` and `tailor` in the background. Call `get_run` with the run id until it finishes. On the CLI, add `--async` and then run `applydiff runs get <run-id> --json`.

## Paste a job description

The job text goes on stdin. Title, company, and URL are flags.

```bash
applydiff add --title "Staff Engineer" --company "Acme" --url "https://example.com/jobs/1" --json <<'EOF'
<paste the job description here>
EOF
```

MCP: `add_job` with the same fields and the description in the tool arguments.

Use the job id from the result for the next commands.

## Draft edits

Pick one path.

**Built-in model.** Run `applydiff tailor <key> --json`. MCP: `tailor`, then `get_run`.

**You draft the edits.** Run `applydiff tailor <key> --context --json` (MCP: `tailor_context`). Read that context. Change only fields that are already there. Do not add a number, metric, or skill that is not in the master CV or `facts.json`. Then save the edits:

```bash
applydiff propose <key> --hunks - --json <<'EOF'
{
  "hunks": [
    {
      "field_id": "summary",
      "section": "summary",
      "before": "text already on the CV",
      "after": "shorter text that matches the job",
      "rationale": "why this field changed",
      "evidence": ["quote from the CV or facts file"]
    }
  ]
}
EOF
```

MCP: `propose_hunks` with the same JSON.

## Hand off

Run `applydiff open <key>` so the user can accept or reject each hunk. If you cannot open a browser, call MCP `review_link` and give them the URL.

After they approve the hunks, they run these themselves:

```bash
applydiff build <key>
applydiff apply <key> --mode manual --channel board
```
