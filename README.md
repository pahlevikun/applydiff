<p align="center">
  <img src="assets/logo-animated.svg" alt="ApplyDiff — git diff for your CV" width="100%" />
</p>

# <img src="assets/icon.svg" alt="" width="28" height="28" align="top" /> ApplyDiff

**git diff for your CV.**

You tailor your CV for every job. Then you can't remember what you changed, you
can't compare versions, and next time you start from the same generic file.
Local-first AI job search, CV tailoring, and interview prep in a single binary (`CLI` + `TUI` + `Web`). Scan job boards with zero model tokens, score postings against your structured CV, review every AI edit **hunk by hunk** before it touches your PDF, and track applications and interview loops in one SQLite file on your machine.

[Quick start](#quick-start) · [How it works](#how-it-works) · [How to use](#how-to-use) · [Comparison](#how-applydiff-compares) · [CLI reference](#cli-commands)

<!-- DEMO GIF: scan -> diff -> approve -> build. 15s, no narration. -->

## Why

Most AI job tools optimize for volume: rewrite the whole resume, auto-fill the form, send 300 applications into silence. You lose track of what changed, you can't verify whether a metric was hallucinated, and next time you start over from the same generic file.

ApplyDiff optimizes for **control and provenance**:

- **Every application is a branch.** Your canonical master CV (`master.json`) stays untouched. Each job gets an isolated branch off master.
- **Every edit is a diff hunk.** The model proposes targeted field-level hunks (`@@ summary @@`, `@@ exp.gofood.b1 @@`, `@@ skills.backend @@`) with rationale and supporting facts—never a full-document rewrite you have to proofread from scratch.
- **Every hunk is approved by you.** Accept, reject, or edit individual hunks in the CLI, keyboard TUI, or Web UI. Only approved hunks compile into `output/<date>-<company>-<role>.pdf`.
- **Every claim is gated by `facts.json`.** A deterministic facts gate blocks any number or metric that does not exist in your verified facts file.

<p align="center">
  <img src="assets/editorial-why-comparison.svg" alt="Editorial comparison: Branching field hunks vs. whole-document AI rewrites" width="100%" />
</p>

### What ApplyDiff will never do

- **Never auto-submit an application.** `applydiff apply` records your submission and channel after you apply; you always click Submit.
- **Never fabricate numbers or skills.** Missing requirements are surfaced explicitly as `gap`, never keyword-stuffed into your CV.
- **Never scrape behind logins.** Sources use public ATS APIs and RSS/JSON feeds with bounded quotas and deduplication.
- **Never phone home.** No account, no cloud sync, no telemetry. Your CV, branches, hunks, and tracker live in one local `applydiff.db` SQLite file.

---

## How it works

<p align="center">
  <img src="assets/editorial-pipeline-convergence.svg" alt="Editorial convergence diagram: Zero model tokens are spent until a posting survives three deterministic gates" width="100%" />
</p>

### Three surfaces, one core

| Surface | Command | Best for |
|---|---|---|
| **CLI** | `applydiff <command>` | Scripting, cron scans, Unix pipes (`--json`, `--dry-run`) |
| **TUI** | `applydiff tui` | Reviewing diffs and running timed interview drills keyboard-only |
| **Web** | `applydiff serve` | Side-by-side paper CV preview, inline hunk controls, full pipeline view |

All three surfaces call the same Go runtime in `internal/` and share the same SQLite database (`applydiff.db`). There is no business logic in the UI.

---

## Quick start

```bash
git clone https://github.com/pahlevikun/applydiff
cd applydiff

# Build one self-contained binary (embeds the Nuxt 3 web UI + SQLite migrations)
make web && make cli

# Import your structured master CV (and sibling facts.json)
./bin/applydiff cv import examples/cv/master.json

# Scan job boards, review in TUI, or open the Web UI
./bin/applydiff scan
./bin/applydiff tui
./bin/applydiff serve     # http://localhost:8080
```

Want a one-command guided walkthrough on sample data? Run:

```bash
./scripts/demo.sh
```

No API key is required when you already have an AI coding CLI logged in (`claude`, `codex`, `cursor-agent`, or `hermes`). PDF compilation uses `xelatex` locally or the bundled Docker image.

---

## How to use

### 1. Import your Master CV and verified facts

ApplyDiff uses a structured JSON master CV alongside a `facts.json` ledger of verified metrics and claims. When you import `master.json`, ApplyDiff automatically loads a sibling `facts.json` when present:

```bash
./bin/applydiff cv import examples/cv/master.json
```

Any number proposed by the LLM during tailoring or STAR story generation that does not appear in your master CV or `facts.json` is rejected by the deterministic facts gate.

### 2. Configure your AI provider (Zero API key or API key)

ApplyDiff auto-detects installed coding-agent CLIs on your `PATH` before falling back to an API key:

**Option A — Coding-agent CLI (recommended, zero API key config):**

If `claude` (`claude -p`), `codex` (`codex exec`), `cursor-agent`, or `hermes` is on your `PATH`, ApplyDiff uses it automatically. You can also pin a specific CLI command:

```bash
export LLM_PROVIDER=cli
export LLM_CLI="claude -p"   # or "codex exec", "cursor-agent"
```

**Option B — OpenAI-compatible API endpoint:**

```bash
export LLM_PROVIDER=api
export LLM_API_KEY="sk-..."
export LLM_MODEL="gpt-4o-mini"                           # default: gpt-4o-mini
export LLM_API_URL="https://api.openai.com/v1/chat/completions"  # or Ollama / OpenRouter
```

**Option C — Offline deterministic mode (for tests and CI):**

```bash
export LLM_PROVIDER=fake
```

### 3. Configure job sources and filters (`jobs/config.yaml`)

Edit `jobs/config.yaml` to enable sources, add Greenhouse company tokens, and tune **hard** vs. **soft** filters:

- **Hard filters (`hard: true`)** skip non-matching listings with an explicit reason (e.g. `regions: [SEA, EU, global]`, `arrangement: [remote, hybrid]`, `min_seniority: [senior]`, `max_posted_days: 14`).
- **Soft filters (`hard: false, boost: N`)** boost matching listings up the feed (`visa_sponsorship`, `eor_friendly`, `async_friendly`, `weekly_days: ["4", "4.5"]`, `salary_min`, `must_have_keywords`).

Preview planned HTTP requests without writing to SQLite, or scan with a preset:

```bash
./bin/applydiff scan --dry-run
./bin/applydiff scan --source greenhouse
./bin/applydiff scan --preset aggressive --max-risk 30
```

> The `aggressive` preset keeps only fresh postings (`<= 5 hours` old) with low competition.

### 4. Inspect jobs, scam flags, and requirement matches

List scanned jobs with their filter scores and decisions, inspect a posting, and compute a requirement-by-requirement match against your master CV:

```bash
# List all scanned jobs (add --json for scripting)
./bin/applydiff jobs

# View job description, deterministic risk/scam flags, and match breakdown
./bin/applydiff show <job-key>

# Score the posting against your master CV (literal / synonym / gap per requirement)
./bin/applydiff match <job-key>
```

### 5. Tailor a branch and review hunks (CLI, TUI, or Web)

Generate a tailored branch proposal for a job and inspect the unified diff:

```bash
# Propose field-level hunks for this job
./bin/applydiff tailor <job-key>

# View colored unified diff in the terminal
./bin/applydiff diff <job-key>

# Accept or reject specific hunks (1-based) or approve all
./bin/applydiff approve <job-key> --hunk 1,2
./bin/applydiff reject <job-key> --hunk 3
./bin/applydiff approve <job-key> --all
```

Prefer reviewing interactively?

- **In the TUI (`./bin/applydiff tui`)**:
  - `Tab` / `Shift+Tab` — switch screens (`Feed`, `Job`, `Diff`, `Sources`, `Applications`, `Interview`)
  - `j` / `k` or `↑` / `↓` — move across jobs or diff hunks (`g` / `G` for top/bottom)
  - `/` — live-filter jobs by title, company, or region
  - `Enter` / `o` — open selected job details
  - `a` / `x` / `Space` — accept hunk, reject hunk, or toggle hunk selection on the `Diff` screen
  - `b` — build the LaTeX PDF for the active job branch
  - `?` — toggle keyboard help overlay (`q` to quit)
- **In the Web UI (`./bin/applydiff serve`)**:
  - Open `http://localhost:8080` to review inline `+`/`-` marks directly on the rendered CV sheet, inspect ATS keyword tags, and click **Accept** / **Reject** / **Edit** per hunk.

### 6. Build the PDF and track your application

Once all hunks on a branch are resolved (no pending hunks remain), compile the tailored PDF and log your application:

```bash
# Compile output/<YYYY-MM-DD>-<company>-<role>.pdf
./bin/applydiff build <job-key>

# Record your manual application and acquisition channel
./bin/applydiff apply <job-key> --mode manual --channel referral --referrer "@alice"

# Measure reply and interview conversion rates per channel
./bin/applydiff stats
```

### 7. Prepare for the interview

Each tracked job has a structured interview kit: predicted interview loop, timeboxed 6-step system design drills, timed mocks with a 5-dimension rubric, and a facts-checked STAR story bank:

```bash
./bin/applydiff interview kit <job-key>      # Generate per-job interview loop & prep kit
./bin/applydiff interview drill <job-key>    # Run the 6-step timeboxed system design drill
./bin/applydiff interview mock <job-key>     # Start a timed mock session
./bin/applydiff interview story <job-key>    # Draft/view facts-verified STAR stories
./bin/applydiff interview score <job-key>    # Score mock performance across 5 dimensions
./bin/applydiff interview export <job-key>   # Export interview prep pack
```

### 8. Export or migrate your database

Your entire state lives in a single SQLite file. Move it across machines with:

```bash
./bin/applydiff export backup.db
./bin/applydiff import backup.db
```

---

## Features

| Area | What ApplyDiff ships |
|---|---|
| **Zero-Token Discovery** | Built-in adapters for **Greenhouse**, **Remotive**, **RemoteOK**, **Jobicy**, **We Work Remotely**, and **Working Nomads**. Deduplicates by normalized URL and posting ID; checks liveness before spending a single LLM token. |
| **First-Class Filters** | Hard (skip with reason) and soft (score boost) filters for work arrangement, 4-day week, region/bloc, timezone overlap, seniority, compensation, visa sponsorship, EOR/contractor friendliness, and posting freshness. |
| **Deterministic Scam Screen** | Rules-based risk engine flags email-domain mismatches, pay-to-work/equipment fees, ID-before-interview requests, clone-and-run malware traps, and missing company domains. Flags warn you; they never silently hide a listing. |
| **Requirement-Level Match** | Scores postings against your structured CV and classifies every requirement as `literal`, `synonym`, or `gap` with evidence. |
| **Hunk-by-Hunk Tailoring** | Proposes atomic field edits (`{ field_id, before, after, rationale, evidence[] }`) enforced by the `facts.json` gate and ATS keyword checker. |
| **LaTeX PDF & Proof Page** | Compiles clean, ATS-parseable PDFs via `xelatex` (`output/<date>-<company>-<role>.pdf`) plus a one-page portfolio proof sheet from the same verified facts. |
| **Channel Funnel Analytics** | Tracks `draft` → `ready` → `applied` across channels (`board`, `referral`, `inbound`, `linkedin`, `email`) so reply and interview rates per channel are measurable via `applydiff stats`. |
| **Interview Prep Suite** | Per-job interview loop prediction, 6-step timeboxed system design framework, 5-dimension rubric scoring, and a STAR story bank that rejects unverified metrics. |

---

## How ApplyDiff compares

ApplyDiff builds on ideas pioneered by [`career-ops`](https://github.com/career-ops-hq/career-ops), [`ai-job-search`](https://github.com/MadsLorentzen/ai-job-search), and [`job-sequencer`](https://github.com/chndranndr/job-sequencer), while solving the core pain point none of them address: **per-hunk diff approval on a branched CV**.

| Capability | [ApplyDiff](https://github.com/pahlevikun/applydiff) | [career-ops](https://github.com/career-ops-hq/career-ops) | [ai-job-search](https://github.com/MadsLorentzen/ai-job-search) | [job-sequencer](https://github.com/chndranndr/job-sequencer) |
|---|---|---|---|---|
| **Approval granularity** | **Per-hunk (`accept` / `reject` / `edit` each line change)** | Whole document | Whole document (drafter-reviewer) | Whole document (`Drafting` → `Ready`) |
| **CV version model** | **Master JSON + isolated branch per job** | Markdown `cv.md` | LaTeX `.tex` files | JSON profile + revision artifacts |
| **Surfaces included** | **CLI + TUI + embedded Web UI in 1 binary** | Agent slash commands + Go TUI | Agent slash commands (`/setup`, `/apply`) | React Web UI + Fastify API |
| **Runtime footprint** | **Single Go binary + pure-Go SQLite** | Node.js + Playwright + Go | Python + Bun + Claude Code | Node.js 24 + Bun + Vite |
| **Facts / anti-hallucination gate** | Deterministic `facts.json` number & claim gate | `cv:verify-facts` script | Drafter-reviewer prompt + checklist | Claim validator + factual-auditor agent |
| **Scam & legitimacy guard** | Deterministic risk rules + score (`0-100`) | Block G posting legitimacy check | Prompt-level untrusted-input rules | Same-run provenance + host allowlist |
| **PDF engine** | LaTeX (`xelatex`) + ATS text checks | HTML + Playwright Chromium | LaTeX (`lualatex` / `xelatex`) + `pdftotext` | LaTeX (`lualatex` / `xelatex`) + `pdfinfo` |
| **Auto-submit** | Never (manual by default) | Never (`You press Submit`) | Never | Never |

---

## Install

**Single binary.** Build once on a machine with Go and Node, then copy `bin/applydiff` anywhere. No Go or Node runtime needed on the target machine.

```bash
make web && make cli
./bin/applydiff serve
```

**Docker (Web UI + API + CLI + TeX Live in one image).**

```bash
docker build -t applydiff .
docker run -p 3000:3000 -v ~/.applydiff:/data applydiff
```

**Release tarballs.** After `make web`, run `./scripts/release.sh` (or `make release`) for cross-compiled binaries under `dist/` plus `sha256sums.txt`.

---

## CLI commands

```text
applydiff cv import <master.json>   Import structured master CV (and sibling facts.json)
applydiff scan                      Pull jobs from configured sources (--source, --preset, --min-risk, --max-risk)
applydiff jobs                      List scanned jobs with filter scores and keep/skip decisions
applydiff show <job-key>            Show job details, risk flags, and requirement match breakdown
applydiff match <job-key>           Compute requirement match score against master CV
applydiff tailor <job-key>          Propose field-level tailoring diff hunks (LLM)
applydiff diff <job-key>            Render unified diff per hunk (--color auto|always|never)
applydiff approve <job-key>         Accept hunks (--hunk 1,2 or --all)
applydiff reject <job-key>          Reject hunks (--hunk 3)
applydiff build <job-key>           Compile LaTeX PDF for an approved branch into output/
applydiff apply <job-key>           Record application submission (--channel, --referrer, --mode)
applydiff stats                     Show reply and interview conversion rates per channel
applydiff interview <subcommand>    kit | mock | drill | story | score | list | export
applydiff tui                       Launch the interactive Bubble Tea terminal UI
applydiff serve                     Launch the embedded Web UI and HTTP API (--port 8080)
applydiff export <file>             Export SQLite database snapshot to a file
applydiff import <file>             Import SQLite database snapshot from a file
```

Global flags: `--json` (structured JSON stdout for scripts/pipes) · `--dry-run` (preview without mutating SQLite).

Exit codes: `0` ok · `2` no match · `3` needs approval · `4` LLM error · `5` LaTeX error.

## Environment

| Variable | Purpose |
|---|---|
| `DATABASE_PATH` | SQLite file path (default `applydiff.db`) |
| `JOBS_CONFIG` | Job sources & filters YAML (default `jobs/config.yaml`) |
| `PORT` | Default HTTP port for `applydiff serve` (default `8080`) |
| `OUTPUT_ROOT` | PDF output directory (default `output/`) |
| `FACTS_PATH` | Verified facts JSON path override (default `facts.json`) |
| `LLM_PROVIDER` | `cli` (auto-detected), `api`, or `fake` (offline deterministic) |
| `LLM_CLI` | Custom coding-agent command when `LLM_PROVIDER=cli` (e.g. `claude -p`, `codex exec`) |
| `LLM_API_KEY` | API key when `LLM_PROVIDER=api` |
| `LLM_MODEL` | Model identifier when `LLM_PROVIDER=api` (default `gpt-4o-mini`) |
| `LLM_API_URL` | Chat completions URL when `LLM_PROVIDER=api` |

---

## Project structure

```text
applydiff/
├── cmd/applydiff/          # Unified CLI, HTTP API server, and Bubble Tea TUI entrypoint
├── internal/
│   ├── source/             # Job board adapters (greenhouse, remotive, remoteok, jobicy, wwr, workingnomads)
│   ├── scan/               # Bounded scanner orchestration & deduplication
│   ├── filter/             # Hard/soft filter evaluator & presets (e.g. aggressive)
│   ├── risk/               # Deterministic scam & legitimacy rule engine
│   ├── match/              # CV requirement matcher (literal / synonym / gap)
│   ├── tailor/             # LLM hunk generator, parser, facts gate, and ATS checker
│   ├── interview/          # Interview loop kit, 6-step drill timer, rubric, STAR story bank
│   ├── store/              # Pure-Go SQLite persistence & timestamped SQL migrations
│   ├── webembed/           # Embedded Nuxt 3 static build served by `applydiff serve`
│   └── runtime/            # Shared application service layer (used by CLI, TUI, and Web)
├── web/                    # Nuxt 3 frontend (Jobs, Match, Diff, My CV, Interview, Sources)
├── jobs/config.yaml        # Declarative sources, search queries, and filter rules
├── examples/cv/            # Sample structured master.json and verified facts.json
└── scripts/                # demo.sh walkthrough and release.sh cross-compiler
```

## Development & verification

Verify the entire codebase offline without live LLM credentials or external database services:

```bash
make test       # Run all unit, migration, filter, risk, TUI snapshot, and API tests (go test ./...)
make web        # Build the Nuxt 3 frontend into internal/webembed/public
make cli        # Build bin/applydiff with embedded assets
make docker     # Build the self-contained Docker image (UI + API + TeX)
make release    # Cross-compile release tarballs + sha256sums.txt into dist/
```

Migrations are timestamped SQL files in `internal/store/migrations/`.

## License

MIT.
