#!/usr/bin/env bash
# End-to-end integration smoke for ApplyDiff (Phase A). Uses a throwaway DB under mktemp.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

OFFLINE=0
KEEP=0
VERBOSE=0
KEEP_DIR=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --offline) OFFLINE=1 ;;
    --keep) KEEP=1 ;;
    --verbose) VERBOSE=1 ;;
    -h|--help)
      echo "Usage: $0 [--offline] [--keep] [--verbose]"
      exit 0
      ;;
    *)
      echo "unknown flag: $1" >&2
      exit 1
      ;;
  esac
  shift
done

log() {
  if [[ "$VERBOSE" == 1 ]]; then
    echo "$@"
  fi
}

if [[ "$KEEP" == 1 && -n "${SMOKE_DIR:-}" ]]; then
  TMP="$SMOKE_DIR"
else
  TMP="$(mktemp -d)"
fi
export SMOKE_DIR="$TMP"
export SMOKE_TMP="$TMP"

mkdir -p "$TMP/output"
export DATABASE_PATH="$TMP/smoke.db"
export OUTPUT_ROOT="$TMP/output"
export MASTER_CV_JSON="$TMP/master_cv.json"
export MASTER_CV_TEX="$TMP/master_cv.tex"
export FACTS_PATH="$TMP/facts.json"
export JOBS_CONFIG="$ROOT/testdata/smoke/jobs.yaml"
export LLM_PROVIDER=fake
export LLM_FAKE_RESPONSE="$(cat "$ROOT/testdata/smoke/tailor-hunks.json")"
export APPLYDIFF_NONINTERACTIVE=1

BIN="$ROOT/bin/applydiff"

log "==> build bin/applydiff"
go build -o "$BIN" ./cmd/applydiff

log "==> import example CV"
"$BIN" cv import "$ROOT/examples/cv/master.json"

SCAN_ARGS=(scan --source remotive)
if [[ "$OFFLINE" == 1 ]]; then
  export REMOTIVE_FIXTURE="$ROOT/internal/source/remotive/testdata/jobs_min.json"
  export APPLYDIFF_SKIP_JOB_LIVENESS=1
  unset REMOTIVE_FIXTURE_URL 2>/dev/null || true
else
  unset REMOTIVE_FIXTURE APPLYDIFF_SKIP_JOB_LIVENESS 2>/dev/null || true
fi

log "==> scan"
SCAN_OUT="$("$BIN" "${SCAN_ARGS[@]}")"
log "$SCAN_OUT"

"$BIN" jobs --json >"$TMP/jobs.json"
JOB_COUNT="$(python3 - <<'PY'
import json, os
with open(os.path.join(os.environ["SMOKE_TMP"], "jobs.json")) as f:
    rows = json.load(f)
print(len(rows))
for row in rows:
    dk = row.get("job", {}).get("dedupe_key") or row.get("dedupe_key")
    if not dk:
        raise SystemExit("job missing dedupe_key: " + json.dumps(row))
if len(rows) < 1:
    raise SystemExit("expected at least one job")
PY
)"

JOB_KEY="$(python3 - <<'PY'
import json, os
with open(os.path.join(os.environ["SMOKE_TMP"], "jobs.json")) as f:
    rows = json.load(f)
if not rows:
    raise SystemExit("no jobs to select")
rows.sort(key=lambda r: r.get("filter", {}).get("score", 0), reverse=True)
print(rows[0]["job"]["id"])
PY
)"

log "==> match $JOB_KEY"
MATCH_OUT="$("$BIN" match "$JOB_KEY")"
log "$MATCH_OUT"

log "==> inbox (needs-decision)"
"$BIN" inbox --json --state needs-decision >"$TMP/inbox.json"
INBOX_OK="$(python3 - <<'PY' "$JOB_KEY"
import json, os, sys
key = sys.argv[1]
with open(os.path.join(os.environ["SMOKE_TMP"], "inbox.json")) as f:
    data = json.load(f)
keys = {row["job_key"] for row in data.get("data", [])}
if key not in keys:
    raise SystemExit(f"job {key} not in needs-decision inbox; keys={keys}")
print(data.get("meta", {}).get("counts", {}).get("needs_decision", 0))
PY
)"

log "==> tailor $JOB_KEY"
TAILOR_OUT="$("$BIN" tailor "$JOB_KEY")"
log "$TAILOR_OUT"
HUNK_COUNT="$(python3 - <<'PY' "$BIN" "$JOB_KEY"
import json, subprocess, sys
bin, key = sys.argv[1], sys.argv[2]
raw = subprocess.check_output([bin, "diff", key, "--json"], text=True)
prop = json.loads(raw)
hunks = prop.get("hunks") or []
if len(hunks) < 1:
    raise SystemExit("expected at least one hunk")
print(len(hunks))
PY
)"

log "==> approve all hunks"
"$BIN" approve "$JOB_KEY" --all

log "==> build $JOB_KEY"
set +e
BUILD_OUT="$("$BIN" build "$JOB_KEY" 2>&1)"
BUILD_RC=$?
set -e
log "$BUILD_OUT"

ARTIFACT=""
BUILD_NOTE=""
if [[ "$BUILD_RC" == 0 ]]; then
  ARTIFACT="$(python3 - <<'PY' "$BUILD_OUT" "$OUTPUT_ROOT"
import glob, json, os, sys
out = sys.argv[1].strip()
root = sys.argv[2]
if out.startswith("{"):
    data = json.loads(out)
    path = data.get("cv_pdf") or data.get("CVPDF") or ""
else:
    path = out.strip()
if path and os.path.isfile(path):
    print(path)
    raise SystemExit(0)
pdfs = sorted(glob.glob(os.path.join(root, "*.pdf")))
if pdfs:
    print(pdfs[-1])
    raise SystemExit(0)
raise SystemExit("build exited 0 but no PDF was written")
PY
)"
elif [[ "$BUILD_RC" == 5 ]]; then
  if ! grep -qiE 'xelatex|tectonic|latex' <<<"$BUILD_OUT"; then
    echo "build exit 5 but output did not mention LaTeX binary:" >&2
    echo "$BUILD_OUT" >&2
    exit 1
  fi
  BUILD_NOTE="latex-missing"
else
  echo "build failed (exit $BUILD_RC):" >&2
  echo "$BUILD_OUT" >&2
  exit "$BUILD_RC"
fi

echo ""
echo "=== smoke summary ==="
echo "temp_dir:      $TMP"
echo "jobs:          $JOB_COUNT"
echo "inbox_needs:   $INBOX_OK"
echo "hunks:         $HUNK_COUNT"
echo "job_key:       $JOB_KEY"
if [[ -n "$ARTIFACT" ]]; then
  echo "artifact:      $ARTIFACT"
else
  echo "artifact:      (none — $BUILD_NOTE)"
fi
echo "offline:       $OFFLINE"
echo "scan:          $SCAN_OUT"

if [[ "$KEEP" != 1 ]]; then
  rm -rf "$TMP"
else
  echo "kept temp dir (--keep): $TMP"
fi
