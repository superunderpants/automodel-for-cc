# automodel-for-cc

Auto-approval guard for Claude Code. **One binary, 3 lines of config, 80% fewer prompts.**

Think of it as the open-source alternative to Claude Code's **auto mode** — self-hosted, works with any LLM provider, no Anthropic API required.

## Why

Claude Code's built-in auto mode is great, but it requires Anthropic API + Sonnet 4.6. If you use a third-party API proxy or different models, you can't enable it. Existing open-source permission tools require extensive manual configuration: custom action types, YAML policies, path whitelists, command rules...

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

## License

MIT

---

# automodel-for-cc 中文文档

## 这是什么

一个 Claude Code 的自动权限审核工具。**一个 exe，三行配置，省掉 80% 的弹窗。**

Claude Code 自带的 auto mode 需要 Anthropic API + Sonnet 4.6。如果你用的第三方 API（如 DeepSeek），这个功能直接不可用。已有的开源方案如 nah 功能强大但配置繁琐：42 种 action type、YAML 策略、路径白名单、命令分类表……

`automodel-for-cc` 反其道而行：**内置安全基线，用户只需配置 LLM API key**。

## 工作原理

```
Claude Code 执行工具前
        ↓
   settings.json hook → automodel-for-cc.exe
        ↓
   第 1 层 — 硬编码白名单（只读工具 + 安全 Bash 命令）
      → 毫秒级放行，不调 API
   第 2 层 — 项目内安全操作（项目目录内的写入/编辑）
      → 毫秒级放行，不调 API
   第 3 层 — 其余所有操作 → LLM 审核
      → 安全就放行，可疑就弹窗（附带 AI 理由）
```

**推理盲（reasoning-blind）**：审核模型只看你的消息和待执行命令，不看 assistant 的推理过程。agent 没法用"这个操作很安全"之类的话术说服审核模型。

## 为什么用 Go

单 exe，零运行时依赖。原生 UTF-8 不乱码，系统证书不报 SSL 错，10ms 启动，不受 Python 环境折腾。

## 快速开始

### 1. 下载

从 [Releases](https://github.com/superunderpants/automodel-for-cc/releases) 页面下载对应平台的二进制文件。

### 2. 配置

创建 `%APPDATA%\auto-guard\config.yaml`（Linux/macOS：`~/.config/auto-guard/config.yaml`）：

```yaml
llm:
  provider: "deepseek"       # deepseek | openai | openrouter | ollama | custom
  api_key: "sk-xxx"          # 你的 API key
  model: "deepseek-chat"
```

三行，完事。

### 3. 挂载 Hook

在 `~/.claude/settings.json` 中添加：

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

重启 Claude Code 即可生效。

### 4. 验证

随便跟 Claude Code 说句话，弹窗大幅减少就是成功了。

## 配置示例

```yaml
# %APPDATA%\auto-guard\config.yaml
llm:
  provider: "deepseek"       # 内置 base_url 映射
  api_key: "sk-xxx"          # 你的 API key
  model: "deepseek-chat"

# 可选：高风险命令跳过 AI 直接弹窗
dangerous:
  - git push --force
  - rm -rf /
```

支持的 provider：`deepseek` | `openai` | `openrouter` | `ollama` | `anthropic` | `custom`

选 `custom` 时才需要手动填 `base_url`。
