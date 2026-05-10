#!/usr/bin/env bash
# automodel-for-cc installer (macOS / Linux)
# Run: curl -sL https://raw.githubusercontent.com/superunderpants/automodel-for-cc/master/install.sh | bash
set -euo pipefail

REPO="superunderpants/automodel-for-cc"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ] && ARCH="amd64"

BIN_NAME="automodel-for-cc-${OS}-${ARCH}"
INSTALL_DIR="${HOME}/.local/bin"
BINARY_PATH="${INSTALL_DIR}/automodel-for-cc"
CONFIG_DIR="${HOME}/.config/auto-guard"
CONFIG_PATH="${CONFIG_DIR}/config.yaml"
SETTINGS_PATH="${HOME}/.claude/settings.json"

echo "=== automodel-for-cc installer ==="

# 1. Download binary
echo "[1/3] Downloading binary..."
mkdir -p "$INSTALL_DIR"
curl -fsSL -o "$BINARY_PATH" "https://github.com/${REPO}/releases/latest/download/${BIN_NAME}"
chmod +x "$BINARY_PATH"
echo "       -> $BINARY_PATH"

# 2. LLM config — 3 fields, same as configuring Claude Code
echo "[2/3] Configuring LLM review (Anthropic API)..."
echo "       Press Enter to use the default shown in brackets."
echo ""

read -rp "Base URL (use Anthropic endpoint): " base_url
base_url="${base_url:-https://api.deepseek.com/anthropic}"

read -rp "API Key: " api_key

read -rp "Model: " model
model="${model:-deepseek-chat}"

# Write config.yaml
mkdir -p "$CONFIG_DIR"
cat > "$CONFIG_PATH" << YAMLEOF
# automodel-for-cc config — same 3 fields as Claude Code
llm:
  base_url: "${base_url}"
  api_key: "${api_key}"
  model: "${model}"
YAMLEOF
echo "       -> $CONFIG_PATH"

# 3. Hook
echo "[3/3] Setting up Claude Code hook..."

if [ -f "$SETTINGS_PATH" ]; then
    SETTINGS=$(cat "$SETTINGS_PATH")
else
    SETTINGS='{}'
fi

add_hook() {
    python3 -c "
import json, sys
settings = json.loads(sys.stdin.read())
settings.setdefault('hooks', {}).setdefault('PreToolUse', [])
hook = {
    'matcher': '.*',
    'hooks': [{'type': 'command', 'command': '$BINARY_PATH'}]
}
pretool = settings['hooks']['PreToolUse']
replaced = False
for i, entry in enumerate(pretool):
    cmds = [h.get('command', '') for h in entry.get('hooks', [])]
    if any('automodel-for-cc' in c for c in cmds):
        pretool[i] = hook
        replaced = True
        break
if not replaced:
    pretool.append(hook)
print(json.dumps(settings, indent=2))
" <<< "$SETTINGS" > "$SETTINGS_PATH"
}

if command -v python3 &>/dev/null; then
    add_hook
    echo "       -> Hook added to $SETTINGS_PATH"
else
    echo "       -> Python3 not found. Add this to ${SETTINGS_PATH}:"
    echo ""
    echo '  "hooks": {'
    echo '    "PreToolUse": [{'
    echo '      "matcher": ".*",'
    echo '      "hooks": [{'
    echo '        "type": "command",'
    echo "        \"command\": \"$BINARY_PATH\""
    echo '      }]'
    echo '    }]'
    echo '  }'
fi

echo ""
echo "Done! Restart Claude Code and you're all set."
echo "Logs: ${CONFIG_DIR}/guard.log"
