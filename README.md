# automodel-for-cc

Auto-approval guard for Claude Code. **One binary, 3 lines of config, 80% fewer prompts.**

Think of it as the open-source alternative to Claude Code's **auto mode** — self-hosted, works with any LLM provider, no Anthropic API required.

## Why

Claude Code's built-in auto mode is great, but it requires Anthropic API + Sonnet 4.6. If you use a third-party API proxy or different models, you can't enable it. Existing open-source tools like nah are powerful but require extensive configuration: 42 action types, YAML policies, path whitelists, command classification tables...

`automodel-for-cc` takes the opposite approach: **sensible defaults, zero config beyond your LLM API key.**

## How it works

```
Your Claude Code session
        ↓
   settings.json hook → automodel-for-cc.exe
        ↓
   ┌── Tier 1: hardcoded allowlist (read-only tools, safe bash)
   │   → instant allow, no API call
   ├── Tier 2: project-safe operations (write/edit inside project dir)
   │   → instant allow, no API call  
   └── Tier 3: everything else → your LLM reviews it
       → returns allow (silent) or ask (prompt with AI reasoning)
```

**Reasoning-blind**: the reviewer LLM sees your messages + the pending command, NOT the assistant's reasoning. So the agent can't talk its way past the guard.

## Quick start

### 1. Download

```bash
# Download the binary for your platform
# Windows: automodel-for-cc.exe
# macOS/Linux: automodel-for-cc
```

### 2. Config

Create `%APPDATA%\auto-guard\config.yaml` (Windows) or `~/.config/auto-guard/config.yaml` (macOS/Linux):

```yaml
llm:
  provider: "deepseek"       # deepseek | openai | openrouter | ollama | custom
  api_key: "sk-xxx"          # your API key
  model: "deepseek-chat"     # model name
```

That's it. Three lines.

### 3. Hook

Add to `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": ".*",
      "hooks": [{
        "type": "command",
        "command": "C:/path/to/automodel-for-cc.exe"
      }]
    }]
  }
}
```

Restart Claude Code. Done.

### 4. Verify

Type anything in Claude Code. You'll see far fewer permission prompts.

## Supported providers

| Provider | Config value |
|----------|-------------|
| DeepSeek | `"deepseek"` |
| OpenAI | `"openai"` |
| OpenRouter | `"openrouter"` |
| Ollama (local) | `"ollama"` |
| Anthropic | `"anthropic"` |
| Custom (any OpenAI-compatible) | `"custom"` + `base_url` |

## Comparison

| | automodel-for-cc | nah | Claude Code auto mode |
|---|---|---|---|
| Setup | 1 binary + 3 lines | pip + 50+ lines config | built-in |
| Config philosophy | sensible defaults | fully programmable | fixed policy |
| Reviewer context | reasoning-blind | full transcript | reasoning-blind |
| API dependency | any OpenAI-compatible | OpenRouter | Anthropic only |
| Latency (tier 1-2) | <1ms | ~50ms (Python) | 0ms |
| Latency (tier 3) | 2-5s | 3-5s | 0.5-2s |

## Build from source

```bash
go build -o automodel-for-cc.exe .
```

Single binary, zero runtime dependencies. Not even a config file parser library at runtime (embedded YAML is compiled in).

## License

MIT
