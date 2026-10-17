# Visual brief: Web UI pipeline (Jobs → Match → Applications → Diff)

## Evidence pack

| Claim | Status |
|-------|--------|
| **Jobs** lists `GET /api/jobs` (feed only); Scan is `POST /api/scan` | verified — `web/pages/jobs.vue`, `httpserver.go` |
| **Match** needs saved master CV (`POST /api/match/…` → 404 without `cv_master`) | verified — `match_handlers.go`, `web/pages/cv.vue` |
| Job routes use `?job=` query (ids contain URL slashes + dedupe bytes) | verified — `web/composables/useJobRoute.ts` |
| **Applications** lists only rows in `applications` (`INNER JOIN`) | verified — `internal/store/application.go` `ListApplications` |
| Successful **Match** creates/updates `applications.status = matched` | verified — `RecordMatchPipeline` in `match_handlers.go`, `decision.go` |
| **Applications** `All` = INNER JOIN only (not every scanned job) | verified — `ListApplications` |
| Pipeline tabs: Matched / Shortlisted / Tailoring / Ready / With CV diff | verified — `pipeline_handlers.go`, `applications.vue` |
| Shortlist / Tailor / Skip on Match call `PUT /api/jobs/:id` → `SetJobDecision` | verified — `match_handlers.go` |
| Tailor POST calls `RecordTailorStarted` → status tailoring + proposal → With CV diff | verified — `tailor_handlers.go` |
| **Diff** loads `GET /api/tailor/:id`; **Tailor** runs LLM then opens Diff | verified — `DiffJobWorkspace.vue`, `MatchJobWorkspace.vue` |

## Takeaway

> After five seconds, the viewer should understand that **Jobs** is the feed, **Match** writes `matched` into the pipeline table, **Applications** tabs filter by `applications.status` (not the whole scan list), and **With CV diff** needs a tailor proposal.

## Visual job + layout

- **Job:** Map relationships and flow across top nav tabs.
- **Family:** **Progression** (left-to-right) with a **binary split**: feed (`jobs` table) vs pipeline (`applications` table).

## Platform

- **Dimensions:** 1200×720 SVG (README, docs).
- **Production:** Deterministic SVG; tokens aligned with `assets/editorial-scan-web-flow.svg`.

## Overlay copy

- Headline: *Jobs is the feed — Applications tabs filter the pipeline table by status*
- Sub: *Match writes matched. Shortlist / Tailor / Export advance status. With CV diff needs a tailor proposal.*

## Asset path

`assets/editorial-web-pipeline-flow.svg`

## QA

| Check | Result |
|-------|--------|
| Clear visual job | pass |
| Evidence in footer | pass |
| Distinct from scan (#04) and token funnel (#02) | pass — nav tabs + DB tables |
| Fake UI | none (diagram only) |

## Limitations

- Does not enumerate every application status (`submitted`, `followup`, etc.).
- CLI/TUI use the same store; only Web nav labels are shown.

## Related

- Scan button: `assets/editorial-scan-web-flow.svg`
- Token funnel: `assets/editorial-pipeline-convergence.svg`
