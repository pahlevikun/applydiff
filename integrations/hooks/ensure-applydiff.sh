#!/usr/bin/env bash
set -euo pipefail
if command -v applydiff >/dev/null 2>&1; then
  exit 0
fi
BIN_DIR="${APPLYDIFF_HOME:-$HOME/.config/applydiff}/bin"
if [[ -x "$BIN_DIR/applydiff" ]]; then
  export PATH="$BIN_DIR:$PATH"
  exit 0
fi
echo "applydiff: not on PATH; install via scripts/install.sh or brew install applydiff" >&2
exit 0
