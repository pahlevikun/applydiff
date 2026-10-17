#!/usr/bin/env bash
# Deprecated for publishing: tag v* and let GitHub Actions run GoReleaser (.goreleaser.yaml).
# Kept for local cross-compiles without GoReleaser (make release).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ ! -d internal/webembed/public ]] || [[ ! -f internal/webembed/public/index.html ]]; then
  echo "release: run make web first to embed the UI" >&2
  exit 1
fi

DIST="$ROOT/dist"
rm -rf "$DIST"
mkdir -p "$DIST"

VERSION="${VERSION:-dev}"
LDFLAGS="-s -w -X main.version=$VERSION"

build_one() {
  local goos="$1" goarch="$2"
  local out="$DIST/applydiff-${goos}-${goarch}"
  [[ "$goos" == "windows" ]] && out="${out}.exe"
  echo "building $out"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -ldflags "$LDFLAGS" -o "$out" ./cmd/applydiff
}

for goos in linux darwin windows; do
  for goarch in amd64 arm64; do
    build_one "$goos" "$goarch"
  done
done

(
  cd "$DIST"
  sha256sum applydiff-* > sha256sums.txt
)

echo "release artifacts in $DIST"
