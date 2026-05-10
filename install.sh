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
SETTINGS_PATH="${HOME}/.claude/settings.json"

echo "=== automodel-for-cc installer ==="

# 1. Download binary
echo "[1/3] Downloading binary..."
mkdir -p "$INSTALL_DIR"
curl -fsSL -o "$BINARY_PATH" "https://github.com/${REPO}/releases/latest/download/${BIN_NAME}"
chmod +x "$BINARY_PATH"
echo "       -> $BINARY_PATH"

# 2. Provider & API key
echo "[2/3] Configuring LLM provider..."
echo ""
echo "  [1] DeepSeek"
echo "  [2] OpenAI"
echo "  [3] Anthropic"
echo "  [4] OpenRouter"
echo "  [5] Ollama (local, no API key needed)"
echo "  [6] Custom (any OpenAI-compatible API)"
echo ""

read -rp "Choose provider [1-6]: " choice
choice="${choice:-0}"

provider=""
env_key=""
case "$choice" in
  1) provider="deepseek";   env_key="DEEPSEEK_API_KEY" ;;
  2) provider="openai";     env_key="OPENAI_API_KEY" ;;
  3) provider="anthropic";  env_key="ANTHROPIC_API_KEY" ;;
  4) provider="openrouter"; env_key="OPENROUTER_API_KEY" ;;
  5) provider="ollama";     env_key="" ;;
  6) provider="custom";     env_key="AUTO_GUARD_API_KEY" ;;
  *) provider="deepseek";   env_key="DEEPSEEK_API_KEY" ;;
esac

echo "       -> Provider: $provider"

if [ "$choice" = "5" ]; then
  echo "       -> Ollama doesn't need an API key (local)"
elif [ "$choice" = "6" ]; then
  read -rp "Enter base URL: " base_url
  echo "export AUTO_GUARD_BASE_URL=\"$base_url\"" >> "${HOME}/.bashrc"
  read -rp "Enter API key (or press Enter to skip): " key
  if [ -n "$key" ]; then
    echo "export AUTO_GUARD_API_KEY=\"$key\"" >> "${HOME}/.bashrc"
  fi
else
  read -rp "Enter API key (or press Enter to skip): " key
  if [ -n "$key" ]; then
    echo "export ${env_key}=\"$key\"" >> "${HOME}/.bashrc"
  fi
fi

if [ "$provider" != "deepseek" ]; then
  echo "export AUTO_GUARD_PROVIDER=\"$provider\"" >> "${HOME}/.bashrc"
fi

echo "       -> Added to ~/.bashrc (source ~/.bashrc or restart terminal)"

# 3. Hook
echo "[3/3] Setting up Claude Code hook..."

# Read or create settings.json
if [ -f "$SETTINGS_PATH" ]; then
    SETTINGS=$(cat "$SETTINGS_PATH")
else
    SETTINGS='{}'
fi

# Merge hook using Python (available on most systems) or node, or fall back to manual
add_hook() {
    python3 -c "
import json, sys
settings = json.loads(sys.stdin.read())
settings.setdefault('hooks', {}).setdefault('PreToolUse', [])
hook = {
    'matcher': '.*',
    'hooks': [{'type': 'command', 'command': '$BINARY_PATH'}]
}
# Replace existing auto-guard entry or append if not found
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
echo "To verify: type anything in Claude Code — you should see fewer permission prompts."
