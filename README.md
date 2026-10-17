# ApplyDiff

**git diff for your CV.**

You tailor your CV for every job. Then you can't remember what you changed, you
can't compare versions, and next time you start from the same generic file.

ApplyDiff treats each application like a branch.

<!-- DEMO GIF: scan -> diff -> approve -> build. 15s, no narration. -->

## Why

Most AI job tools optimize for volume. Send more, faster. You end up with 300
applications, no replies, and no memory of what you actually sent.

ApplyDiff optimizes for control:

- Every edit is a diff.
- Every diff is approved by you, hunk by hunk.
- Every application keeps its own version of your CV, so you can see exactly
  what went out.

The model drafts. You decide. Then it builds the PDF.

## Quick start

```bash
git clone https://github.com/pahlevikun/applydiff
cd applydiff

# one binary: CLI + TUI + web UI + API
make web && make cli
./bin/applydiff scan
./bin/applydiff tui
./bin/applydiff serve     # http://localhost:8080
```

No API key needed when you already have a coding-agent CLI logged in
(`claude`, `codex`, or `cursor-agent`).

PDF builds need `xelatex`, or Docker with the bundled image.

## Three surfaces, one core

| Surface | Command | Use it for |
|---|---|---|
| **CLI** | `applydiff <command>` | scripting, cron, pipes (`--json`, `--dry-run`) |
| **TUI** | `applydiff tui` | reviewing diffs fast, keyboard only |
| **Web** | `applydiff serve` | seeing the whole picture |

All three call the same Go core in `internal/` and share one SQLite database.
There is no business logic in the UI.

## What it does

- **Scan** job boards and drop duplicates and dead listings. Sources ship for
  Greenhouse, Remotive, RemoteOK, Jobicy, We Work Remotely, and Working Nomads.
- **Match** each listing against your structured CV, with reasons, not just a
  score.
- **Tailor** with line-level edits. Not a rewrite you have to proofread from
  scratch.
- **Review** each change as a hunk. Accept, reject, or edit. Same behavior in
  the web UI, the terminal, and the CLI.
- **Build** a LaTeX PDF from the approved version, into
  `output/<date>-<company>-<role>.pdf`.
- **Screen for scams** with deterministic rules: email domain mismatch,
  pay-to-work, ID before interview, clone-and-run tests, contact with no
  company domain. Hard flags mark the job, they never hide it.
- **Track** each application: draft, ready, applied, plus the channel
  (board, referral, inbound, LinkedIn) so reply rate per channel is measurable.
- **Prepare for the interview** with a per-job kit: the likely loop, the six
  timeboxed system design steps, a timed mock with a five-dimension rubric, and
  a STAR story bank that cannot use numbers your CV can't back up.
- **Export** the SQLite file and move it to another machine.

Filters are first-class: work arrangement, region, timezone overlap, seniority,
compensation, visa, freshness. Hard filters skip with a reason. Soft filters
boost. A preset called `aggressive` keeps only jobs posted in the last five
hours with low competition.

## Safety

- **No auto-submit.** Applying is a manual step by default.
- Assisted and auto modes are opt-in, and auto stays limited to email and ATS
  APIs.
- No browser bots. Nothing behind a login gets scraped.
- Every action is logged with actor, time, and the hunks involved.
- The scam screen is a filter, not a guarantee. You always decide.

## Privacy

Local-first. Your CV, your applications, and your history live in one SQLite
file on your machine. No account, no telemetry. Move it with
`applydiff export backup.db`.

## Stack

Go for scanning and matching. SQLite for storage, so your data is one file.
Bubble Tea for the TUI. Nuxt 3 for the web, embedded into the binary. LaTeX for
the CV, because ATS still reads it best.

## Install

**Single binary.** Build once on a machine with Go and Node, then copy
`bin/applydiff` anywhere. No Go, no Node at runtime.

```bash
make web && make cli
./bin/applydiff serve
```

**Docker (UI + API + TeX in one image).**

```bash
docker build -t applydiff .
docker run -p 3000:3000 -v ~/.applydiff:/data applydiff
```

**Release tarballs.** After `make web`, run `scripts/release.sh` for
cross-compiled binaries under `dist/` plus `sha256sums.txt`.

## CLI commands

`scan`, `jobs`, `show`, `match`, `tailor`, `diff`, `approve`, `reject`, `build`,
`apply`, `stats`, `interview` (kit, mock, drill, story, score, list, export),
`export`, `import`, `serve`, `tui`

Exit codes: `0` ok · `2` no match · `3` needs approval · `4` LLM error ·
`5` LaTeX error.

## Environment

| Variable | Purpose |
|---|---|
| `DATABASE_PATH` | SQLite file (default `applydiff.db`) |
| `JOBS_CONFIG` | Job sources YAML (default `jobs/config.yaml`) |
| `PORT` | Default HTTP port (default `8080`) |
| `OUTPUT_ROOT` | PDF output directory (default `output/`) |
| `FACTS_PATH` | JSON facts file (default `facts.json`) |

## Development

```bash
make cli        # build bin/applydiff
make web        # build the Nuxt UI into the embed dir
make test       # go test ./...
make pdf        # print the PDF build usage
make docker     # build the image
make release    # cross-compiled binaries into dist/
```

Migrations are timestamped SQL files in `internal/store/migrations/`.

## Palette

Phosphor: amber CRT on warm black. Tokens live in
`web/assets/tokens-root.css` and `internal/theme/theme.css` (TUI).

## License

MIT.
