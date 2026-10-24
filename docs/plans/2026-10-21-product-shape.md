# ApplyDiff — Product shape, IA, and UX

Status: plan · Owner: F Y Pahlevi · Date: 2026-10-21

## 1. Problem

The app is a full web app with six destinations: Jobs, Match, Diff, CV,
Interview, Sources. It works, but the shape is wrong for the job it does.

Three concrete failures:

1. **The user has to know the workflow.** Nothing says "start here, then here".
   Diff is a separate page, so people only find it if they already know it exists.
2. **Nothing has initiative.** The app waits to be opened. Job hunting is
   repetitive and easy to forget, so an app that waits gets forgotten for days.
3. **No state for work in progress.** Five applications half-decided look
   identical to zero applications. There is no queue, only lists.

This is not a styling problem. Repainting the current screens will not fix it.

## 2. Decision

Three layers, one source of data.

```
ENGINE (daemon)        starts on its own, no user present
  scheduled scan · dedupe · match · verdict · scam screen · draft tailoring
  -> produces a QUEUE OF DECISIONS

COCKPIT (where humans decide)   TUI + Web over the same API
  see the queue · review diff · approve · build · send · prep

ON-RAMP (thin agent skill)      wraps the CLI --json, no logic of its own
```

Rejected shapes and why:

- **Agent-only (Claude/Cursor as the whole product).** Great for distribution,
  terrible for the core loop: approving 20 hunks in a chat transcript has no
  stable list, no keyboard, no side-by-side. Chat is a bad diff viewer.
- **App-only.** Correct for review, wrong for initiative. Pull-only products
  get forgotten.
- **Extension-first.** The decision surface needs width. A 400px popup cannot
  hold a diff review. Extension is a capture tool, later.
- **Desktop (Tauri) first.** Cosmetics. The bottleneck is not "hard to open".

## 3. Information architecture

### 3.1 One queue, four states

The home screen is not a list of jobs. It is a queue of decisions.

| State | Meaning | Primary action |
|---|---|---|
| **Needs decision** | new job passed filters, draft ready | Review |
| **In progress** | tailoring or hunks unfinished | Continue |
| **Sent** | applied, waiting for a reply | Track |
| **Interview** | reply received, prep not done | Prepare |

Everything else (Jobs, Match, Diff) becomes a *detail view of a queue item*,
not a top-level destination. The user never has to know where to go.

### 3.2 Five destinations, not six

```
Inbox        queue of decisions (default)
Pipeline     all applications, kanban by status
CV           master CV (rarely touched)
Interview    prep for active applications
Settings     sources, filters, automation, notifications
```

Match and Diff are tabs inside an Inbox item. That is the whole fix.

### 3.3 Item detail

One item = one job = one screen with four tabs:

```
[ Fit ] [ Diff ] [ Letter ] [ Interview ]
```

- **Fit** — score, five dimensions with written reasons, requirements
  (match / partial / gap with the CV field that proves it), legitimacy flags,
  freshness verdict, risk flags.
- **Diff** — the only full-width screen in the product. Document flows freely;
  change marks attach inline; accept/reject buttons attach to the hunk.
- **Letter** — cover letter draft, diffed against the previous one.
- **Interview** — kit, loop, topics, questions to ask.

Bottom bar: decisions made so far, and one primary button
(`Build PDF` / `Continue` / `Send`).

## 4. Screen specs

### 4.1 Inbox (default)

- Left: filters as chips (preset, freshness verdict, risk, region,
  arrangement). Collapsed by default; one active filter visible.
- Center: queue items, grouped by state, most urgent first.
- Each card: title, company, region, age, verdict badge, risk badge, and
  **exactly one primary button**.
- Keyboard: `j/k` move, `space` open, `x` skip (with reason), `g/G` ends.

Target: decide 40 items in under 3 minutes.

### 4.2 Diff review (the hero)

- Left rail: CV sections with hunk counts.
- Center: the document, running text, `ins`/`del` inline.
- Right of each hunk: `acc` / `rej` buttons.
- Bottom: `12 of 18 decided`, `Build` disabled until every hunk has a decision
  or an explicit skip.
- Keyboard: `j/k` hunk, `a` accept, `x` reject, `e` edit, `space` toggle,
  `b` build.
- Must survive leaving and coming back: decisions persist per hunk, not per
  document.

### 4.3 Pipeline

Columns: Needs decision · In progress · Sent · Interview · Closed.
Card shows company, role, days since sent, silence state.
Drag is optional; keyboard move is required.

### 4.4 CV

Structured editor, not a textarea. Every field has an ID, because diffs
reference field IDs, not strings. Fields whose numbers are absent from the
facts file are visibly marked.

### 4.5 Interview

Kit (loop + topics) on the left, the active step on the right: timebox,
required output, probes. Mock timer, rubric 1-5, story bank with empty
categories flagged.

### 4.6 Settings

Sources (enabled, last scan, last error), filters, automation caps,
notifications, data export.

## 5. States that must exist

| State | Requirement |
|---|---|
| Empty | teaches the next action, never just "no data" |
| Loading | skeleton shaped like the content, no generic spinner |
| Partial | source A failed, source B succeeded: show B, name A's failure |
| Error | per source, with retry on that source only |
| Stale | last scan older than 24h: say so on every screen |
| Offline | daemon not running: say how to start it |

## 6. Daemon

`applydiff daemon` runs until stopped.

- Ticks: scan every N hours (default 6), match after each scan, draft for jobs
  above the score threshold, cap drafts per day.
- All work writes a `runs` row: kind, started, finished, fetched, stored,
  skipped, error.
- If a source fails, the run succeeds partially and records the failure per
  source. One dead board never kills a run.
- Max concurrency: one scan at a time. A second invocation waits or exits
  cleanly with a message.
- Graceful shutdown on SIGINT/SIGTERM: finish the current source, write state.
- Logs to file, not stdout only.
- No LLM call without a configured backend; degraded mode = scan + match only.

## 7. Notifications

- Telegram (cheapest, we already have the bot): one digest per run,
  "N applications need your decision", with a link to Inbox.
- Desktop notification via the daemon when the web UI is open.
- Never notify for scans that found nothing.

## 8. Automation policy

| Fully automatic | Needs a human | Never |
|---|---|---|
| scan on schedule | approve hunks | submit without explicit opt-in |
| dedupe, first_seen, verdict | choose to apply | send email without confirmation |
| match scoring | manual edits | answer a recruiter |
| scam screen | tone of the cover letter | claim experience that does not exist |
| draft tailoring (not sending) | | |

Caps: drafts only for jobs above threshold, max N per day (default 10).
Every automated write is visible in the run log.

## 9. Data model additions

- `jobs.first_seen` (set on insert, never updated)
- `jobs.role_key` (company + normalized title, location clause removed)
- `jobs.verdict` (fresh / stale / re-stamped / likely-evergreen / unknown)
- `jobs.removed_at` (absent from the last successful board crawl)
- `queue_items` (job_id, state, reason, updated_at) — or derive from
  applications; derive first, materialize only if slow
- `runs` (already exists) gains per-source error detail
- `notifications` (sent_at, kind, payload) to avoid double-sending

## 10. API and CLI

HTTP (local):
```
GET  /api/inbox?state=
GET  /api/jobs/{key}
POST /api/jobs/{key}/decision       {action: shortlist|skip, reason}
GET  /api/jobs/{key}/diff
POST /api/jobs/{key}/hunks/{n}      {decision: accept|reject|edit, text}
POST /api/jobs/{key}/build
POST /api/jobs/{key}/apply          {mode}
GET  /api/runs?limit=
```

CLI:
```
applydiff daemon [--interval 6h] [--drafts 10] [--notify telegram]
applydiff inbox [--state needs-decision] [--json]
applydiff review <job-key>          # opens the TUI diff screen directly
```

## 11. Phases

### Phase 1 — daemon + queue (this PR)
- [ ] `applydiff daemon` with tick loop, run log, graceful shutdown
- [ ] scheduled scan + match + optional draft (cap enforced)
- [ ] derived inbox query: needs-decision / in-progress / sent / interview
- [ ] `GET /api/inbox`, `applydiff inbox --json`
- [ ] tests: tick runs once, cap respected, second daemon exits cleanly

### Phase 2 — Inbox UI
- [ ] Web `/` becomes Inbox (queue), not the job list
- [ ] item detail with four tabs
- [ ] one primary action per card
- [ ] all six states from section 5

### Phase 3 — diff persistence + keyboard
- [ ] per-hunk decisions persist across navigation
- [ ] full keyboard map in Web and TUI
- [ ] progress counter + build gate

### Phase 4 — notifications
- [ ] Telegram digest per run
- [ ] dedupe via `notifications`

### Phase 5 — agent on-ramp
- [ ] thin skill wrapping the CLI, no logic
- [ ] MCP server exposing the same four calls

## 12. Acceptance for the whole plan

- A new user can go from clone to "here are 5 decisions waiting" without
  reading documentation.
- Every screen has one obvious primary action.
- Leaving mid-review and returning loses nothing.
- A failed source is visible, never silent.
- No automated send, ever, without an explicit per-job opt-in.

## 13. Non-goals

- Multi-user, accounts, hosted SaaS.
- Browser bots or scraping behind logins.
- Desktop wrapper before the queue exists.
- Extension before the queue exists.
