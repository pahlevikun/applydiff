# Upstream feature inventory

Full feature list from the three reference repos, down to the concrete detail.
Used to decide what ApplyDiff borrows (patterns only) and what it adds.

---

## 1. career-ops (JS/Node, MIT, ~74k★)

Local-first, AI-agnostic job-search agent that runs in an AI coding CLI. Logic
lives in Markdown prompt files (`modes/`) + Node scripts; no model is hardcoded.

### Discovery & sources
- **106 job-board providers** (`providers/*.mjs`): Greenhouse, Ashby, Lever,
  Workable, BambooHR, Breezy, Comeet, Built In, ADP, Amazon, Alibaba, and more.
- **Plugin system** (`plugins/` + `plugins-registry/`): hooks `provider` /
  `ingest` / `search` / `notify` / `export`. No auto-submit hook exists.
  Trust badges (bundled / approved / community-unverified / off-registry).
  Tamper detection: plugin hash is checked; changes without a version bump
  block the plugin until re-pinned.
- **No-token scan**: portal scan, liveness check, and PDF generation use zero
  model tokens. `scan --verify` drops dead postings before they cost anything.
- **Dedupe**: `normalizePostingUrl(url)` and `aggregatorPostingId(url)` for
  LinkedIn/Indeed-style multi-employer aggregators. Same key = same row.
- **Batch processing**: `batch-input.tsv` → orchestrator → N headless CLI
  workers, with progress state, retries, and resume.

### Evaluation (the "A-H" report)
One report per posting, scored 1-5 across 5 weighted dimensions:
- A. Role summary
- B. CV match, per requirement: importance + gaps + mitigation
- C. Level strategy
- D. Company research (WebSearch)
- E. CV personalization plan
- F. Interview prep (STAR stories)
- G. Posting legitimacy (scam / ghost-job signals)
- H. Draft application answers — only when score ≥ 4.5

### Tailoring & output
- CV is `cv.md` → HTML template → PDF via Playwright (not LaTeX here).
- `latex-content.mjs` / `latex-escape.mjs` exist for a LaTeX path.
- Verification scripts: `cv:verify-facts` (no fabricated numbers), `cv:verify-ats`,
  `jd:similarity`, `cover-letter`, `pdf`.

### Application (apply mode)
- **ATS auto-fill**: reads the form in the user's browser, drafts answers,
  types text, selects dropdowns.
- **CRITICAL: never submits. You click Submit.** Checkboxes, radio, and
  captcha always stay with the user; skipped fields are listed with values.
- Per-ATS quirks handled:
  - Ashby — duplicate-candidate merge check, suggests `you+team@email.com`
  - Lever — hCaptcha on checkboxes, so text/select only
  - Workable — SPA stale DOM, clipboard paste + fresh query per field
  - react-select — type char-by-char, re-snapshot DOM

### Tracking & pipeline integrity
- Tracker = `data/applications.md` + TSV merges (`merge-tracker.mjs`).
- `verify-pipeline.mjs` (statuses, duplicates, links), `dedup-tracker.mjs`,
  `normalize-statuses.mjs`, `fix-report-links.mjs`, `cv-sync-check.mjs`.

### Plugins & safety
- Bundled plugins: gmail, notion, obsidian, google-calendar, outlook-interviews,
  serper, tavily, apify, docx, markdown, linkedin-alerts, startup-boards,
  commute, h1b-sponsor.
- Network safety: `ctx.fetch` only, manifest `allowedHosts` enforced, SSRF
  protection, hostname validation, IP guard, DNS cache, trust validator.
- No arbitrary shell; sources return `Job[]`, the engine writes them.

### UI
- Dashboard TUI (Go): filter tabs (All / Evaluated / Applied / Interview /
  Top≥4 / No-Apply), sort (score/date/company/status), grouped/flat, lazy
  report preview, inline status picker.
- Web app (Next.js).

### Principles
- Local-first, no account, no server in the loop.
- Human-in-the-loop: prepares + evaluates, the human reviews and clicks.
- Data contract: strict system-files vs user-files separation; the updater
  never touches user data.

---

## 2. ai-job-search (Python + Claude Code, MIT, ~46k★)

"Fork it and own it." A drafter-reviewer pipeline built on Claude Code, with
LaTeX output and per-portal search CLIs.

### Commands (slash workflow)
`/setup` (profile) → `/scrape` (portals) → `/rank` → `/apply <url>` →
`/interview` → `/outcome`. Plus `/add-portal`, `/add-template`, `/html-report`,
`/gmail-sync`, `/notion-sync`, `/upskill`, `/reset`, `/expand`.

### Drafter-reviewer (the core pattern)
- Drafter writes the CV + cover letter.
- Reviewer agent runs in a **fresh context** (drafts passed inline, not
  re-read), researches the company, and returns a critique.
- Critique is split into:
  - Part A — **structured edits** as JSON `{file, old_string, new_string}`,
    applied mechanically with the Edit tool. Skip any that would fabricate.
  - Part B — narrative suggestions (company angles, reframing), applied with
    judgment after verifying each company claim via WebSearch.
- Verification checklist runs exactly once, at the end.

### LaTeX output
- CV compiles with `lualatex`; cover letter with `xelatex` (cover.cls needs
  fontspec). Mandatory PDF compilation in `/apply`.
- `verify_pdf` + `verify_layout`: text extraction, reading order matches
  visual order, no interleaved sections. Custom templates may swap the
  compiler (e.g. `typst`), never fall back blindly.

### ATS keyword coverage
- Extract required/preferred keywords from the JD once, match each against the
  extracted PDF text, in the posting's language.
- Classify: **literal** / **synonym-only** / **missing (gap)**.
- Gap is left missing, never keyword-stuffed. Honesty rule.

### Application forms (third artifact, `08-application-forms.md`)
For portal free-text fields beyond CV/cover letter:
- Self-introduction (100-200 words, lead with evidence, tie to employer).
- Structured project entries (name/role/dates/description, scope honesty,
  60-word short variant).
- Hard character limits (draft 4-6 candidates, count programmatically).
- Every claim must trace to the candidate profile, master CV, or CLAUDE.md.

### Data & safety
- `job_key.py` — stable job identity for dedupe.
- `robots_check.py` — scraping compliance.
- `security_guards.py` — no auto-submit.
- Tracker CSV (`job_search_tracker.csv`) for status + dedupe/exclusion.
- Thin-pointer design: one source of truth (`CLAUDE.md` + `.claude/` +
  `.agents/skills/`) so Claude Code, Codex, Antigravity, Gemini, Cursor all
  read the same specs.

### Portal search CLIs
Per-board TypeScript CLIs under `.agents/skills/` (SKILL.md + CLI + tests):
freehire, jobbank, jobdanmark, jobindex, jobnet, linkedin. Each is a portable
Agent Skill; Codex/Antigravity auto-discover them.

---

## 3. job-sequencer (TS/React/Fastify/SQLite/Pi, NO LICENSE)

The closest full build. Local-first single-user dashboard with bounded agent
workflows and human approval gates.

### Menus (5)
1. **Jobs** — scrape, review, filter, score, select.
2. **Applications** — generate docs, approve, record apply, follow-up.
3. **Interview** — practice for applied jobs.
4. **Profile & Criteria** — candidate facts + search preferences.
5. **Settings** — provider/model, sources, score threshold, test connection.

### Agent pipeline (per job, bounded)
research → strategist → writer → reviser → critic →
claim-validator + factual-auditor (facts gate) → evidence →
ATS reviewer → visual QA.
Each stage emits structured output; the facts gate blocks any number that is
not in the verified profile facts.

### Adaptive search with enforced limits
- 7 built-in sources: FreeHire, LinkedIn, TokyoDev, Japan Dev, Relocate.me,
  Y Combinator Remote, Indeed Indonesia. Custom sources use bounded
  declarative HTTP(S) templates + JSON paths / HTML selectors.
- Quotas: call count, result slots, time, per-source requests. A source that
  yields nothing is deprioritized within the same run.
- Detail fetches require same-run provenance.

### Approval gates (the key difference vs ApplyDiff)
```
Drafting → (explicit approval) → Ready → (explicit manual record) → Applied
```
- Generation stops at Drafting. Approving to Ready is a separate user action.
- Recording Applied is manual (the user submits the application themselves).
- Below-threshold jobs go to **Discarded**, never deleted.

### Evidence-backed generation
Claims in drafts are linked to verified profile facts; deterministic checks
cover explicit evidence rules, model review covers semantic claims, and human
review remains required.

### Safety boundaries
- Loopback only (127.0.0.1). No auth, CORS, cloud sync, job submission, email,
  or automatic provider fallback.
- Pi sessions disable ambient skills/extensions and unrestricted built-ins;
  sources expose typed `searchJobs` / `fetchJobDetails` only. No arbitrary
  commands or user code.
- Credentials stay in Pi auth store or env; never committed.

### UI
- React + Vite frontend (Tracker is the only surface), Fastify API,
  `node:sqlite` (no separate DB service).
- Explicit anti-slop rules: no gradients, no glassmorphism, no rounded-card
  dashboard, no AI sparkles/badges, no rainbow status colors, no generic KPI
  cards, text label over icon when it scans faster.
- Global run strip shows the active AI run across screens.

### Verification
- `npm run check` (whitespace, types, doc links, manifest, architecture
  boundaries), `npm test` (fixtures), `npm run eval` (AI fixtures +
  trajectory evaluation), `smoke:browser`, `smoke:latex`, `doctor`.

---

## What ApplyDiff adds (the differentiator, present in none of the above)

- **Per-hunk diff approval**: each AI edit is a field-level hunk the owner
  accepts / rejects / edits individually, rendered as a unified diff against
  the master CV. The other three approve a whole document, never per change.
- **Master + branch version model**: one canonical structured CV; each job is
  a branch off master. Accepted hunks build that job's PDF; master stays clean;
  drift is flagged stale.
- **Job switcher**: prev/next re-binds the preview, fit score, keyword rail,
  and rationale to the selected job instantly.
- **Proof page (portfolio) artifact**: a third output (identity line + 3 real
  results + links + contact), generated from the same verified facts.
