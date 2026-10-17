# Job sources catalog

ApplyDiff keeps **one merged catalog** of job sources, aligned with:

- **career-ops** — `providers/*.mjs` on [career-ops-hq/career-ops](https://github.com/career-ops-hq/career-ops) (109 board providers; `local-parser` excluded)
- **ai-job-search** — portal CLIs: freehire, jobbank, jobdanmark, jobindex, jobnet, linkedin
- **job-sequencer** — FreeHire, LinkedIn, TokyoDev, Japan Dev, Relocate.me, Y Combinator Remote, Indeed Indonesia
- **career-ops plugins** — linkedin-alerts, startup-boards (ingest paths, not HTML scrape)
- **ApplyDiff roadmap** — Hiring.cafe, Wellfound, Otta, AlignJobs.id, Glassdoor, etc.

**Machine-readable list:** [`internal/jobconfig/sources_catalog.yaml`](../../internal/jobconfig/sources_catalog.yaml) (embedded at build time).

Run `go test ./internal/jobconfig/ -run Catalog` for current counts.

## Scan modes

| Mode | Meaning |
|------|---------|
| `live` | Registered fetcher; `applydiff scan` when enabled in `jobs/config.yaml` |
| `planned` | In catalog; fetcher not wired yet (most career-ops providers) |
| `manual` | Paste URLs / alert links |

## Live today (7)

| ID | Name |
|----|------|
| `greenhouse` | Greenhouse |
| `remotive` | Remotive |
| `remoteok` | Remote OK |
| `jobicy` | Jobicy |
| `weworkremotely` | We Work Remotely |
| `workingnomads` | Working Nomads |
| `fourdayweek` | 4 Day Week |

## Parity notes

- **career-ops `4dayweek`** → catalog id `fourdayweek` (same API).
- **career-ops `jobbankca`** → Job Bank Canada; **ai-job-search `jobbank`** → Danish Jobbank (separate id `jobbank`).
- **LinkedIn** — catalog row for tracking; HTML scrape intentionally not implemented (ToS). Use `linkedin_alerts` + **manual** until alert ingest lands.
- **echojobs** — marked retired upstream; kept in catalog for parity only.

## Adding a fetcher

1. Add or update the row in `internal/jobconfig/sources_catalog.yaml` (`scan: live` when done).
2. Implement `internal/source/<id>/` with `Fetch` + tests.
3. Wire `SourcesConfig`, `registerSources`, `queriesFor`, `BuildFetchPlans`, and `liveConfigToggles` in `sources_catalog.go`.

Regenerate from career-ops provider list (optional):

```bash
# Example: list provider module names from upstream
curl -sL 'https://api.github.com/repos/career-ops-hq/career-ops/contents/providers?per_page=200' \
  | jq -r '.[].name' | grep '\.mjs$' | grep -v '^_' | sed 's/.mjs$//'
```

Merge new ids into `internal/jobconfig/sources_catalog.yaml` with `origins: career_ops`.
