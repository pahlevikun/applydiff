# ApplyDiff — Execution plan (smoke → data correctness → sources → release)

Status: plan · Owner: F Y Pahlevi · Date: 2026-10-24
Supersedes the open items in `2026-10-21-improvements.md` for sequencing only.

## 0. Where we actually are

Done and verified per phase:

- CLI (14 subcommands), TUI (5 screens), Web (embedded), one binary
- `internal/interview` — kit, mock, rubric, grounded story bank
- `internal/risk` — 10 deterministic scam rules + `stats` per channel
- Phase 1-5 of the product shape: daemon + derived inbox, Inbox UI + job
  detail tabs, per-hunk persistence + keyboard + build gate, notifications,
  MCP server + agent skill
- 6 sources: greenhouse + remotive, remoteok, jobicy, weworkremotely,
  workingnomads
- Releases v1.0.0 and v1.1.0, repo public, all commits under the owner

Not in code yet (plan only):

- End-to-end run with real data
- `first_seen`, `role_key`, `re-stamped`, status-never-overwritten
- freshness verdict, liveness 3-state, `applydiff explain`, conditional tier
- ATS adapters beyond greenhouse, ATS detection from a careers URL
- README that describes the current product, demo GIF, v1.2.0 tag

## 1. Rules for every PR in this plan

1. One phase = one branch = one PR. Merge immediately after verification.
2. Every commit authored AND committed by `Farhan Yuda Pahlevi <pahlevi.kun@gmail.com>`,
   zero `Co-authored-by` trailers. Use `/root/.hermes/scripts/pr_finalize.sh`.
3. The finalizer prints deletions vs main. Any deletion must be justified in the
   PR body, otherwise the PR is rejected.
4. Tests must grow. Package count may not shrink.
5. No CI workflows. No palette changes. No auto-submit.
6. Verification is run by us, not taken from the agent's report. The agent's
   claim is a hypothesis until a command proves it.

## 2. Phase A — end-to-end smoke test (do this first)

Why: every phase passed alone, but the pipeline has never run start to finish
with real data. Integration bugs hide exactly here.

Deliverable: `scripts/smoke.sh` plus a Go test that runs the same path offline.

Steps the script performs, on a throwaway DB (`mktemp -d`):

1. Build the binary.
2. Import the example CV (`examples/cv/master.json`, `examples/cv/facts.json`).
3. Scan one live source (default `remotive`; `--offline` uses fixtures).
4. Assert `jobs --json` returns at least 1 row and every row has a `dedupe_key`.
5. Assert `inbox --json` shows the row in `needs_decision`.
6. Tailor one job with the deterministic fake LLM runner (no network, no cost).
7. Assert a diff exists and has at least 1 hunk.
8. Approve every hunk through the CLI.
9. Assert `build` either produces `output/<date>-<company>-<role>.pdf`, or exits
   with code 5 and a message naming the missing LaTeX binary. Both are pass,
   silence is fail.
10. Print a summary with counts and the paths written.

Rules:

- Idempotent: running twice leaves the same DB consistent, no duplicate jobs.
- Never writes to the user's real DB. Always `DATABASE_PATH` inside the temp dir.
- Never calls a paid LLM. Fake runner only.
- `--offline` mode must work with no network at all.

Acceptance:

- `./scripts/smoke.sh` exits 0 on this box.
- `./scripts/smoke.sh --offline` exits 0 with the network unplugged.
- The Go test `TestSmokePipeline` covers steps 4-9 without shelling out.

## 3. Phase B — data correctness (P0 bugs)

These are wrong today; every scan makes them worse.

### B1. `first_seen` and never-overwrite

- Migration: `jobs.first_seen TEXT NOT NULL DEFAULT ''`.
- On insert: set to now (UTC ISO-8601).
- On conflict: **never** update `first_seen`.
- Upsert must stop clobbering user state: `status`, and any decision/note column,
  are left untouched on conflict. Non-identity descriptive fields are updated
  only when the incoming value is non-empty (`CASE WHEN excluded.x != '' ...`).

### B2. `role_key`

- Column `jobs.role_key TEXT NOT NULL DEFAULT ''` + index.
- Definition: `normalizeCompany(company) + '\x00' + normalizeTitle(title)`.
- `normalizeTitle`: lowercase, collapse whitespace, strip HTML tags, decode
  entities, drop a trailing location/qualifier clause after `,` `-` `|` `(`…`)`
  unwrapping parentheses into their contents (keep the words, drop the brackets).
- Dedupe collapses rows sharing `role_key` + same company, keeping the oldest
  `first_seen` as canonical, and records every source in `origins_json`.
- URL stays as the display/apply link; it is no longer part of identity.

### B3. `re-stamped`

- Derived, not stored: `posted_at - first_seen > 7 days` marks a row as
  re-stamped.
- Surfaced as a badge + a filter, never auto-rejected.

Acceptance for Phase B:

- Test: two consecutive scans with a changed `posted_at` keep `first_seen` and
  keep `status`.
- Test: the same role from three different sources yields one row with three
  origins.
- Test: a row whose board date is 40 days newer than first sighting is flagged.
- Existing 32 test packages stay green.

## 4. Phase C — ATS sources (company-direct), three waves

Brief already written: `applydiff-goal-ats-wave1.txt`.

- Wave 1 (15 JSON adapters + `BoardSource` + `atsdetect`)
- Wave 2 (HTML-based: teamtailor, jobvite, softgarden, freshteam, jazzhr, avature)
- Wave 3 (long tail + regional ATS)

Rules:

- Clean-room: read the upstream adapter once for the contract, close it, write
  ours from the contract. No copied lists, no shared names.
- One fixture test per adapter, no network in tests.
- `CREDITS.md` records prior art only (no code copied). If any curated list is
  ever copied, it moves to a licensed section with the MIT notice.

Acceptance per wave:

- `applydiff source add <careers-url> --dry-run` detects the right provider.
- `applydiff scan --board provider/slug` returns plausible rows for at least one
  real board per wave.
- No adapter silently drops a page: partial results come with an error.

## 5. Phase D — honesty features

- Freshness verdict: `fresh / stale / re-stamped / ghost / unknown`, with
  `likely-evergreen` requiring two independent signals (age, repost count,
  mass-posting, curated text marker). One signal is never enough.
- Liveness: `Live / Uncertain / Expired` + strike counter. `Uncertain` never
  advances a close.
- `applydiff explain`: audit configured terms — match count, three examples,
  flag terms matching nothing or eating an unusual share.
- Match honesty: a skill that appears only in a course/learning line scores
  zero; a hidden salary or seniority is `unknown`, not a guess.

## 6. Phase E — README, demo, release

- README rewritten around the queue: "scan runs in the background, you only
  decide". Keep the hook, add the daemon and MCP.
- Demo GIF, 15 seconds, no narration: daemon tick → inbox → one diff → build.
- Tag `v1.2.0`, rebuild the 6 binaries + checksums, publish the release.
- `applydiff doctor` mentions missing LaTeX and missing LLM backend clearly.

## 7. Order and gates

```
A  smoke test          gate: exits 0 online + offline
B  data correctness    gate: 3 new tests, no status clobber
C1 ATS wave 1          gate: 15 fixture tests + 1 live board per provider that allows it
C2 ATS wave 2          gate: same
C3 ATS wave 3          gate: same
D  honesty features    gate: verdict + liveness + explain tests
E  README + v1.2.0     gate: fresh clone builds, GIF present
```

Do not start C before B lands: the ATS adapters write into the same upsert path,
and fixing identity afterwards means re-crawling everything.

## 8. Definition of done for this plan

- `./scripts/smoke.sh` passes on a clean clone.
- A single role posted on three boards appears once.
- A board that lies about its posting date is visibly flagged.
- One command discovers a company's ATS from its careers URL and starts using it.
- README explains the queue in five seconds and shows it moving.
- Every commit under the owner's name, no co-author trailers, no open PRs.
