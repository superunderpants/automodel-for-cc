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

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/superunderpants/automodel-for-cc/master/install.ps1 | iex
```

**macOS / Linux:**
```bash
curl -sL https://raw.githubusercontent.com/superunderpants/automodel-for-cc/master/install.sh | bash
```

One line. Downloads the binary, asks for your API credentials (same 3 fields as Claude Code), and wires up the hook. Restart Claude Code and you're done.

> **Manual install?** See [below](#manual-install).

## Manual install

### 1. Download

**Windows (PowerShell):**
```powershell
$dir = "$env:APPDATA\auto-guard"; mkdir $dir -Force | Out-Null
curl -L -o "$dir\automodel-for-cc.exe" https://github.com/superunderpants/automodel-for-cc/releases/latest/download/automodel-for-cc-windows-amd64.exe
```

**macOS / Linux:**
```bash
mkdir -p ~/.local/bin
curl -sL -o ~/.local/bin/automodel-for-cc https://github.com/superunderpants/automodel-for-cc/releases/latest/download/automodel-for-cc-$(uname -s | tr A-Z a-z)-$(uname -m)
chmod +x ~/.local/bin/automodel-for-cc
```

### 2. Config

Create `config.yaml` — same 3 fields you used to configure Claude Code:

*Windows:* `%APPDATA%\auto-guard\config.yaml`
*macOS/Linux:* `~/.config/auto-guard/config.yaml`

```yaml
llm:
  base_url: "https://api.deepseek.com/anthropic"
  api_key: "sk-xxx"
  model: "deepseek-chat"
```

Uses the Anthropic Messages API format (the same protocol Claude Code uses). Works with any provider that speaks Anthropic API: DeepSeek, SiliconFlow, OpenRouter, Anthropic, or your own proxy.

Alternatively, set environment variables:
```bash
export AUTO_GUARD_BASE_URL="https://api.deepseek.com/anthropic"
export AUTO_GUARD_API_KEY="sk-xxx"
export AUTO_GUARD_MODEL="deepseek-chat"
```

API key fallback chain: `AUTO_GUARD_API_KEY` → `ANTHROPIC_AUTH_TOKEN` → `OPENAI_API_KEY`.

### 3. Hook

Add to `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": ".*",
      "hooks": [{
        "type": "command",
        "command": "/path/to/automodel-for-cc"
      }]
    }]
  }
}
```

Restart Claude Code.

## Config reference

```yaml
# %APPDATA%\auto-guard\config.yaml  (Windows)
# ~/.config/auto-guard/config.yaml   (macOS/Linux)
llm:
  base_url: "https://api.deepseek.com/anthropic"
  api_key: "sk-xxx"
  model: "deepseek-chat"
  timeout: 10          # seconds (optional, default 10)
```

## Uninstall

Remove the binary and the hook entry from `~/.claude/settings.json`:

**Windows:**
```powershell
rm -r $env:APPDATA\auto-guard
# Then manually delete the "PreToolUse" entry referencing automodel-for-cc
# from ~/.claude/settings.json
```

**macOS / Linux:**
```bash
rm ~/.local/bin/automodel-for-cc
rm -rf ~/.config/auto-guard
# Then manually delete the "PreToolUse" entry referencing automodel-for-cc
# from ~/.claude/settings.json
```

## Build from source

```bash
go build -o automodel-for-cc.exe ./src/
```

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

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/superunderpants/automodel-for-cc/master/install.ps1 | iex
```

**macOS / Linux:**
```bash
curl -sL https://raw.githubusercontent.com/superunderpants/automodel-for-cc/master/install.sh | bash
```

一行命令。自动下载、配置 API 信息（和 Claude Code 一样的 3 个字段）、挂载 hook。重启 Claude Code 即可生效。

> 手动安装？见[下方](#手动安装)。

## 手动安装

### 1. 下载

**Windows (PowerShell):**
```powershell
$dir = "$env:APPDATA\auto-guard"; mkdir $dir -Force | Out-Null
curl -L -o "$dir\automodel-for-cc.exe" https://github.com/superunderpants/automodel-for-cc/releases/latest/download/automodel-for-cc-windows-amd64.exe
```

**macOS / Linux:**
```bash
mkdir -p ~/.local/bin
curl -sL -o ~/.local/bin/automodel-for-cc https://github.com/superunderpants/automodel-for-cc/releases/latest/download/automodel-for-cc-$(uname -s | tr A-Z a-z)-$(uname -m)
chmod +x ~/.local/bin/automodel-for-cc
```

### 2. 配置

创建 `config.yaml`，和配置 Claude Code 一样的 3 个字段：

*Windows:* `%APPDATA%\auto-guard\config.yaml`
*macOS/Linux:* `~/.config/auto-guard/config.yaml`

```yaml
llm:
  base_url: "https://api.deepseek.com/anthropic"
  api_key: "sk-xxx"
  model: "deepseek-chat"
```

走 Anthropic Messages API 协议（和 Claude Code 同一个协议）。支持所有兼容 Anthropic API 的服务：DeepSeek、硅基流动、OpenRouter、Anthropic 官方、自建代理等。

也可以用环境变量：
```bash
export AUTO_GUARD_BASE_URL="https://api.deepseek.com/anthropic"
export AUTO_GUARD_API_KEY="sk-xxx"
export AUTO_GUARD_MODEL="deepseek-chat"
```

API key 读取顺序：`AUTO_GUARD_API_KEY` → `ANTHROPIC_AUTH_TOKEN` → `OPENAI_API_KEY`。

### 3. Hook

在 `~/.claude/settings.json` 中添加：

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": ".*",
      "hooks": [{
        "type": "command",
        "command": "/path/to/automodel-for-cc"
      }]
    }]
  }
}
```

重启 Claude Code。

## 配置参考

```yaml
# %APPDATA%\auto-guard\config.yaml  (Windows)
# ~/.config/auto-guard/config.yaml   (macOS/Linux)
llm:
  base_url: "https://api.deepseek.com/anthropic"
  api_key: "sk-xxx"
  model: "deepseek-chat"
  timeout: 10          # 秒（可选，默认 10）
```

## 卸载

删除二进制文件和 `~/.claude/settings.json` 中的 hook 配置即可：

**Windows:**
```powershell
rm -r $env:APPDATA\auto-guard
# 然后手动删除 ~/.claude/settings.json 中引用 automodel-for-cc 的 PreToolUse 条目
```

**macOS / Linux:**
```bash
rm ~/.local/bin/automodel-for-cc
rm -rf ~/.config/auto-guard
# 然后手动删除 ~/.claude/settings.json 中引用 automodel-for-cc 的 PreToolUse 条目
```
