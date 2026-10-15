# ApplyDiff — Implementation Plan

> **For agentic workers:** Implement task by task with `super-tdd` (one failing test, then the smallest pass), prove each task with `super-verify`, commit with `atomic-semantic-commit`. Steps use checkbox (`- [ ]`) syntax.

**Goal:** One local web app that scans job boards, matches each listing against a structured CV, has AI draft tailoring edits, and lets the owner approve those edits per-hunk as diffs before a LaTeX PDF is built.

**Architecture:** Go API + scanner workers, Postgres store, Nuxt 3 web UI, LLM runner that shells to a CLI agent or hits an API directly. The CV lives as structured data; LaTeX is an export target, never the source of truth. Every AI change is a field-level hunk the owner accepts or rejects.

**Tech Stack:** Go 1.22 · PostgreSQL 16 · Nuxt 3 (Vue 3) · TypeScript · latexmk + TeX Live (Docker) · model runner (CLI agent | direct API).

## Global Constraints

- Local-first. Runs on the owner's machine. No account, no server in the loop.
- Human-in-the-loop. The app never submits an application. It never sends email. The owner clicks.
- Structured CV is the source of truth; LaTeX is generated.
- Cheapest model for scan/eval/tailor unless a task needs more.
- No fabricated experience, titles, or metrics. A number that is not in `facts` cannot enter a diff.
- No em dashes, no banned AI glossary (leverage, utilize, robust, synergy) in CV copy.
- One person per workspace (multi-person ready, not multi-person in v1).

---

## Repo layout

```
applydiff/
├── cmd/
│   ├── applydiff/        # api server
│   └── scan/             # scanner worker entrypoint
├── internal/
│   ├── job/              # Job model, dedupe, liveness
│   ├── source/           # one file per source + registry
│   ├── cv/               # structured CV model + facts gate
│   ├── match/            # fit scoring (rubric)
│   ├── tailor/           # draft + reviewer, emits hunks
│   ├── diff/             # hunk engine, apply/revert
│   ├── build/            # LaTeX render + compile + ATS test
│   ├── llm/              # runner interface (cli | api)
│   └── store/            # postgres + migrations
├── web/                  # Nuxt 3
├── docs/
│   ├── plans/
│   └── wireframes/       # s01..s05 HTML + PNG
├── fixtures/             # sample JDs + a sample CV
└── docker-compose.yml    # postgres + texlive
```

**Reference projects (patterns only, no code copied):**
- `career-ops` — source/provider pattern, no-token scan, liveness, dedupe, A-H evaluation, facts + ATS verification, data contract.
- `ai-job-search` — drafter-reviewer, structured edit JSON (`old_string`/`new_string`), LaTeX toolchain, per-portal CLI, `job_key`.

---

## Structured CV schema (the core decision)

The CV is JSON, not LaTeX. Every field has a stable id so a diff can target it.

```jsonc
{
  "person": "farhan-yuda-pahlevi",
  "contact": {
    "name": "Farhan Yuda Pahlevi",
    "headline": "Senior Software Engineer",
    "email": "...", "site": "...", "linkedin": "...", "github": "...",
    "location": "Bogor, West Java, Indonesia",
    "timezone_line": "GMT+7 · available for EU and US Eastern overlap"
  },
  "summary": { "id": "summary", "text": "..." },
  "skills": [
    { "id": "skills.backend", "label": "Backend", "items": ["Go", "Rails", "..."] }
  ],
  "experience": [
    {
      "id": "exp.gofood",
      "company": "GoTo Group", "team": "GoFood Merchant",
      "title": "Senior Software Engineer, Fullstack",
      "location": "Jakarta, Indonesia",
      "start": "2023-03", "end": null,
      "bullets": [ { "id": "exp.gofood.b1", "text": "..." } ]
    }
  ],
  "education": [ { "id": "edu.gunadarma", "school": "...", "degree": "...", "year": "..." } ],
  "achievements": [ { "id": "ach.1", "text": "..." } ]
}
```

`facts.json` holds every approved metric, keyed by id (`abac_impl: "40%"`). A hunk that introduces a number not present in `facts` is blocked by the facts gate.

### Version model: master + branch (per-job CV)

One canonical **master** CV, one **branch** per job, exactly like git.

- `master` is the source of truth. It is edited manually only; tailoring never writes to it.
- A job opens a **branch** = `master` at that point. The AI's hunks are diffs against master.
- You accept/reject hunks on the branch only. Accepted hunks build that job's PDF.
- Master stays clean. Rejecting a hunk on job A never affects job B.
- Every hunk targets a field id, so a later master edit can be detected against open branches and flagged **stale** (like a git conflict) for re-review.

```
master (cv.json)
 ├── job:honor    = master + accepted hunks
 ├── job:covergo  = master + accepted hunks
 └── job:gramian  = ...
```

---

## Phases

### Phase 0 — Scaffold

- [ ] Go module, `cmd/applydiff`, `cmd/scan`, package skeleton
- [ ] `docker-compose.yml`: postgres 16 + texlive image
- [ ] Nuxt 3 app in `web/` with the 5 screens stubbed
- [ ] `make dev`, `make test`, `make scan` targets
- [ ] CI: `go test ./...`, `npm test`, `go vet`, lint

### Phase 1 — Sources and job feed

- [ ] `Job` model + migration (`company`, `title`, `url`, `source`, `posted_at`, `remote`, `region`, `status`)
- [ ] `internal/source` interface: `Fetch(queries) []RawJob`, registry keyed by id
- [ ] Sources: Greenhouse, Lever, Ashby, Workable (the four from the threads), then Hiring.cafe, RemoteOK, We Remotely, Remotive, Wellfound, Otta, YC, Jobicy, LinkedIn (discovery-only)
- [ ] `normalizeURL(url)` + `dedupeKey(job)` (URL + `company+role`)
- [ ] Liveness check drops dead postings before any LLM cost
- [ ] `jobs/config.yaml` mirrored from the latex-resume repo (sources, queries, filters, caps)
- [ ] UI: **s01 Jobs** — feed, filters (fit/region/source), status rail, preview panel
- [ ] Test: fixture scan returns N jobs, dedupe removes known dupes, dead postings dropped

### Phase 2 — Structured CV

- [ ] `cv` package: load `cv.json`, validate against schema, expose `Field(id)`
- [ ] Importer: parse `resume_ats_v2.tex` → `cv.json` (one-shot), plus `facts.json` from AGENTS.md metrics
- [ ] Exporter: `cv.json` → `.tex` fragments → full document → latexmk → PDF
- [ ] ATS text test: extract text from PDF, assert reading order + keyword coverage
- [ ] UI: **s04 Structured CV** — section tree, field editor, provenance, per-field history
- [ ] Test: round-trip tex → json → tex keeps metrics, `facts` gate blocks a fake number

### Phase 3 — Match

- [ ] `match` package: score a job against the CV, 5 dimensions (skills, seniority, domain, logistics, motivation), total 0-25
- [ ] Requirement extraction from JD → per-requirement match/partial/gap with the CV field that supports it
- [ ] Legitimacy check (company real, official ATS URL, salary consistency)
- [ ] Persist score + dimension breakdown + matched fields
- [ ] UI: **s02 Match** — JD mapped to CV side by side, fit bars, decision actions
- [ ] Test: known JD + CV gives the expected band; a genuine gap is labeled gap, not match

### Phase 4 — Tailor + diff approval (the point of the app)

- [ ] `tailor` package: drafter pass produces a candidate CV; reviewer pass critiques (drafter-reviewer pattern)
- [ ] Emit edits as hunks: `{ field_id, section, before, after, rationale, evidence[] }`
- [ ] `diff` package: apply hunk, revert hunk, apply-all-accepted, produce the "after" CV
- [ ] Facts gate on hunk creation; ATS-keyword tag on hunks that add a JD term
- [ ] UI: **s03 Diff** — files rail (sections), unified diff center (red/green with +/− markers), rationale/evidence right, per-hunk Accept/Reject/Edit, bottom action bar
- [ ] Job switcher: prev/next at the top of s03. Switching job swaps the branch instantly — the preview, fit score, keyword rail, and rationale all re-bind to the new job. A job with no tailoring yet shows master as-is plus a "Tailor for this JD" action.
- [ ] Tests: hunk apply/revert is reversible; a hunk with an unapproved number is rejected; accepted-only build matches the accepted set
- [ ] Accept gate: PDF build runs only on accepted hunks; PDF preview in the right rail

### Phase 5 — LLM runner

- [ ] `llm.Runner` interface: `Complete(prompt) string`
- [ ] `cli` runner: shell to a CLI agent (Claude Code, Codex) with the prompt; parse JSON out
- [ ] `api` runner: direct API call, cheapest model default
- [ ] Prompt templates carry the CV schema, the JD, the rubric, and the facts list
- [ ] Test: a canned prompt returns parseable hunks from both runners (mock the CLI)

### Phase 6 — Applications and tracking

- [ ] `application` model: job + CV version + accepted hunks + status
- [ ] Status enum: `discovered · matched · shortlisted · tailoring · ready · submitted · followup · closed`
- [ ] Follow-up rows for already-applied roles (Speechify, RevenueCat from the queue)
- [ ] UI: **s05 Sources/Settings** — sources grid with toggles, queries, filters, LLM runner choice
- [ ] Export: accepted CV → `cv-{company}-{date}.pdf` + cover letter draft
- [ ] Test: an application freezes the CV version it was built from

---

## Design system (UI)

Dark workbench. Tokens in `web/assets/tokens.css`.

- Surface `#0b0e13` / `#11151d` / `#171c26`, line `#232a36`
- Text `#e7eaf0`, muted `#8b95a5`, dim `#5c6675`
- Accent `#3fd0be` (brand), amber `#d9a441` (attention/pending), add `#3fb950`, del `#f0553f`
- Type: IBM Plex Sans (prose), IBM Plex Mono (fields, ids, diff, metrics)
- Radius 10 / 7. Spacing scale 4-6-8-11-14-16-22.
- Diff is a "forge" ledger, not a GitHub clone: 44px mono line-number gutter, +/− markers, section headers as `@@ summary @@`.
- Custom 24px stroke icons only. No lucide. No default clean-stroke set.

Wireframes: `docs/wireframes/s01..s05` (HTML source + PNG).

---

## Tailoring rules (bahan dari riset komunitas)

Rules that shape the drafter prompt, sitting next to the facts gate.

**Non-English source is rewritten, not translated.** When the master CV is Indonesian and the target posting is English, the drafter rewrites into CV-English. It never translates word for word. Prompt fence, verbatim: "Jangan nambahin pengalaman, angka, atau skill yang nggak ada di teks aku."

**Bullets start with a past-tense action verb.** Each experience bullet opens with a verb (Architected, Shipped, Built). Keep a number only if the source has it. Never invent one.

**Mirror the posting's wording, not its claims.** Use the term the posting uses when the concept is already true in the CV. Never add a skill the person does not have. This pairs with the ATS-keyword tag on hunks.

**Verify against the original before export.** Diff the tailored CV against the source point by point: changed meaning, differing numbers, claims absent from the source, grammar. Runs as a gate before the PDF builds.

**Never translate proper nouns.** Company, university, and official job titles stay as written on the official site or LinkedIn. Only descriptions translate.

**Localized prompts live in `internal/tailor/prompts/`.** English plus Indonesian first (Bahasa Indonesia market). Each is a data file, not code.

---

## Open questions

- Honor + Anovium: confirm Indonesia eligibility before they reach the qualified gate.
- Which CLI agent is the default runner on the owner's box?
- Cover letter: same structured diff flow in v1, or Phase 6+?
- Personal access to LinkedIn (cookies) — discovery only, caps stay at 5/day.
