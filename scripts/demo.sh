#!/usr/bin/env bash
# ApplyDiff demo — idempotent local walkthrough (network used only for scan).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DB="${DATABASE_PATH:-$ROOT/demo-applydiff.db}"
export DATABASE_PATH="$DB"
export JOBS_CONFIG="${JOBS_CONFIG:-$ROOT/jobs/config.yaml}"

echo "==> build"
mkdir -p bin
go build -o bin/applydiff ./cmd/applydiff

echo "==> import example CV"
./bin/applydiff cv import examples/cv/master.json

echo "==> scan (live sources)"
./bin/applydiff scan

echo "==> jobs (filter score as risk proxy)"
./bin/applydiff jobs --json | head -c 4000

TOP="$(./bin/applydiff jobs --json | python3 -c "
import json,sys
rows=json.load(sys.stdin)
rows=[r for r in rows if r.get('filter',{}).get('keep')]
rows.sort(key=lambda r: r.get('filter',{}).get('score',0), reverse=True)
print(rows[0]['job']['id'] if rows else '')
" 2>/dev/null || true)"

if [[ -z "$TOP" ]]; then
  echo "No kept jobs after scan; try relaxing filters in jobs/config.yaml"
  exit 0
fi

echo "==> top match: $TOP"
./bin/applydiff match "$TOP" || true

echo "==> tailor (kit) for $TOP"
./bin/applydiff tailor "$TOP" || true

echo ""
echo "Next: ./bin/applydiff diff $TOP && ./bin/applydiff approve $TOP --all && ./bin/applydiff build $TOP"
