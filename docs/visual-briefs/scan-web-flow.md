# Visual brief: Web Scan button flow

## Evidence pack

| Claim | Status |
|-------|--------|
| Jobs **Scan** calls `POST /api/scan` then `GET /api/jobs` | verified — `web/pages/jobs.vue` `runScan()` |
| HTTP handler uses 5-minute context timeout | verified — `cmd/applydiff/httpserver.go` `handleScan` |
| Web scan is **synchronous** (not `Runs.Start` / `--async`) | verified — handler calls `s.app.Runner.Run` directly |
| Pipeline: fetch → cap → dedupe → extract → risk → filter → liveness → upsert | verified — `internal/scan/run.go` |
| No LLM on scan | verified — no `llm` import in scan package |
| UI shows `lastScan` counts (fetched, stored, skipped, dead) | verified — `jobs.vue` template |
| CLI `scan --async` differs from web | verified — `cmd/applydiff/commands.go` only |

## Takeaway

> After five seconds, the viewer should understand that **Scan** is a blocking, zero-LLM fetch-and-filter pass that fills SQLite and then reloads the Jobs list.

## Visual job + layout

- **Job:** Explain the mechanism (click → API → Go pipeline → DB → refresh).
- **Family:** **Progression** (left-to-right) with a **convergence** of sources into one runner; outcome split (stored / skipped / dead).

## Platform

- **Dimensions:** 1200×680 SVG (docs, README, blog).
- **Production:** Deterministic SVG; Phosphor editorial tokens aligned with `assets/editorial-pipeline-convergence.svg`.

## Overlay copy

- Headline: *Click Scan → the server fetches boards, filters in Go, then refreshes your job list*
- Sub: *Synchronous POST (up to 5 min). Same pipeline as `applydiff scan` — no model calls.*

## Asset path

`assets/editorial-scan-web-flow.svg`

## QA

| Check | Result |
|-------|--------|
| Clear visual job | pass |
| Evidence in footer | pass |
| Distinct from pipeline convergence (#02) | pass — web entry, no token threshold to tailor |
| Fake UI | none (diagram only) |

## Limitations

- Does not show per-source error handling UI (errors surface as HTTP 500 + `error` on page).
- Cap numbers not duplicated from `jobs/config.yaml` (see pipeline asset for caps).

## Related

- Full product funnel: `assets/editorial-pipeline-convergence.svg`
- Standalone surfaces: `assets/editorial-standalone-three-ways-in.svg`
