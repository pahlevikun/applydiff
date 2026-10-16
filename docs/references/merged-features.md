# ApplyDiff — merged feature set

The final product: features from career-ops, ai-job-search, and job-sequencer,
merged into one local-first app, plus ApplyDiff's own differentiator. Stack:
Go + Nuxt 3 + SQLite (pure-Go `modernc.org/sqlite`).

Principles that hold across everything: local-first, human-in-the-loop, never
auto-submit, facts-only (a number that is not in the verified facts cannot
enter the app).

---

## 1. Discovery (job search)

Merged from career-ops (provider breadth + plugin) and job-sequencer
(adaptive search + quotas).

- **Sources, in priority order:**
  1. ATS resmi — Greenhouse, Lever, Ashby, Workable
  2. Niche remote boards — Remote OK, We Work Remotely, Working Nomads,
     Remotive, Wellfound, Otta, YC, Jobicy
  3. Aggregators — Hiring.cafe, AlignJobs.id
  4. LinkedIn — discovery only, capped
- **Plugin system** for new sources (hook: provider / ingest / search / export),
  with trust badges and tamper detection.
- **No-token scan** first: portal scan + liveness + PDF all run without model
  cost. Dead postings dropped before any LLM spend.
- **Dedupe**: normalized posting URL + aggregator posting id.
- **Adaptive quotas**: call, result, time, and per-source budgets. A source
  that yields nothing is deprioritized within the same run.
- **Recency**: freshest first, `<24h` tag, optional alert on new matches.

## 2. Match & evaluate

Merged from career-ops (A-H rubric) + job-sequencer (fit score vs profile).

- **A-H report** per posting: role summary, CV match per requirement
  (importance + gaps + mitigation), level strategy, company research,
  personalization plan, STAR interview prep, legitimacy check, draft answers
  (only ≥4.5).
- **Fit score** 1-5 across five dimensions: skills, seniority, domain,
  logistics, motivation.
- **Requirement → CV field mapping**, classified literal / synonym-only / gap.
  Gaps stay gaps, never stuffed.
- **Legitimacy**: company verified, official ATS URL, salary consistency.
- Below threshold → **Discarded** (kept, never deleted).

## 3. Tailor (AI generation)

Merged from ai-job-search (drafter-reviewer) + job-sequencer (full agent
pipeline) + ApplyDiff's facts gate.

- **Agent pipeline** (per job, bounded): research → strategist → writer →
  reviser → critic → claim-validator + factual-auditor → evidence → ATS
  reviewer → visual QA.
- Output = **structured hunks** `{ field_id, before, after, rationale,
  evidence[] }`, each targeted at a field id.
- **Facts gate**: a number not in `facts.json` is blocked.
- **Rewrite, don't translate** (for non-English source): bullets start with a
  past-tense verb; keep numbers only if the source has them; mirror the JD's
  terms, never add a skill the person lacks; never translate proper nouns;
  verify against the source before export.
- **LLM runner**: CLI agent (Claude Code / Codex) or direct API, cheapest
  model by default.

## 4. Diff approval (ApplyDiff's own — the differentiator)

Present in none of the three repos.

- **Per-hunk accept / reject / edit**: every AI change is an individual hunk.
- **Inline full-CV preview**: the CV renders as a scrollable page; changes are
  marked inline (+/− in the margin, no boxes); Accept/Reject buttons sit on the
  changed line.
- **Master + branch**: one canonical master CV; each job is a branch. Accepted
  hunks build that job's PDF. Master stays clean; drift flags stale.
- **Job switcher**: prev/next re-binds the preview, fit score, keyword rail,
  and rationale to the selected job instantly.
- **Rationale + evidence rail**: why each edit, and which fact supports it.

## 5. Build & export

Merged from ai-job-search (LaTeX toolchain) + job-sequencer (visual QA).

- Structured CV → LaTeX → PDF (lualatex for CV; xelatex for cover letter).
- **ATS text test**: extract text, assert reading order + keyword coverage.
- **Visual QA** on the compiled PDF.
- **Proof page** artifact: one-page portfolio (identity line + 3 real results
  + links + contact), from the same verified facts.
- **Cover letter**: same structured-diff flow.

## 6. Application tracking

Merged from job-sequencer (approval gates + tracker) + career-ops (autofill).

- Status machine: Drafting → (approve) → Ready → (record) → Applied →
  follow-up.
- **Autofill assist** (optional, career-ops style): drafts answers and types
  fields, never touches checkboxes/captcha, never submits.
- **Follow-up drafts** + interview-prep chat keyed to the job.

## 7. UI

Merged from job-sequencer (anti-slop) + ApplyDiff's wireframes.

- 5 screens: s01 Jobs, s02 Match, s03 Diff, s04 Structured CV, s05 Sources.
- Dark workbench; IBM Plex Sans + Mono (+ Serif for the CV page); accent teal,
  add/del green/red, pending amber.
- Anti-slop: no gradients, glassmorphism, rounded-card dashboards, AI
  sparkles/badges, rainbow status colors, or generic KPI cards. Text label
  over icon when it scans faster. A component communicates state, enables an
  action, or is dropped.
- Custom 24px icon set. No lucide.

---

## What we deliberately do NOT build

- No auto-submit, no email sending, no cloud sync, no auth, no multi-user.
- No generic workflow engine, no autonomous search-to-apply pipeline.
- No LaTeX-as-source-of-truth (structured JSON is the truth; LaTeX is export).
