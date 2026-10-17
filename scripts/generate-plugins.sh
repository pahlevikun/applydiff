#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SKILL_SRC="$ROOT/integrations/skill/SKILL.md"

mkdir -p "$ROOT/.agents/skills/applydiff"
cp "$SKILL_SRC" "$ROOT/.agents/skills/applydiff/SKILL.md"
mkdir -p "$ROOT/.cursor-plugin/skills/applydiff"
cp "$SKILL_SRC" "$ROOT/.cursor-plugin/skills/applydiff/SKILL.md"
mkdir -p "$ROOT/.claude-plugin/skills/applydiff"
cp "$SKILL_SRC" "$ROOT/.claude-plugin/skills/applydiff/SKILL.md"

cat > "$ROOT/.cursor-plugin/plugin.json" <<'EOF'
{
  "name": "applydiff",
  "version": "0.1.0",
  "description": "ApplyDiff job tailoring with human approval gate"
}
EOF

cat > "$ROOT/.cursor-plugin/mcp.json" <<'EOF'
{
  "mcpServers": {
    "applydiff": {
      "command": "applydiff",
      "args": ["mcp"]
    }
  }
}
EOF

cat > "$ROOT/.claude-plugin/plugin.json" <<'EOF'
{
  "name": "applydiff",
  "version": "0.1.0",
  "description": "ApplyDiff job tailoring with human approval gate"
}
EOF

cat > "$ROOT/.claude-plugin/.mcp.json" <<'EOF'
{
  "mcpServers": {
    "applydiff": {
      "command": "applydiff",
      "args": ["mcp"]
    }
  }
}
EOF

mkdir -p "$ROOT/.cursor-plugin/hooks" "$ROOT/.claude-plugin/hooks"
cat > "$ROOT/.cursor-plugin/hooks/hooks.json" <<'EOF'
{
  "version": 1,
  "hooks": {
    "sessionStart": [{ "command": "./integrations/hooks/ensure-applydiff.sh" }],
    "beforeShellExecution": [{ "command": "./integrations/hooks/deny-gated.sh" }]
  }
}
EOF

cat > "$ROOT/.claude-plugin/hooks/hooks.json" <<'EOF'
{
  "hooks": {
    "SessionStart": [{ "command": "./integrations/hooks/ensure-applydiff.sh" }],
    "PreToolUse": [{ "command": "./integrations/hooks/deny-gated.sh" }]
  }
}
EOF

echo "plugins generated"
