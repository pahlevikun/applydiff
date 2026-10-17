#!/usr/bin/env bash
# Deny shell commands that bypass the human approval gate.
INPUT="${1:-}"
if echo "$INPUT" | grep -Eq 'applydiff[[:space:]]+(approve|reject|build|apply)\b'; then
  echo "ApplyDiff: approve/reject/build/apply must be run by a human in a terminal or the web UI." >&2
  exit 2
fi
exit 0
