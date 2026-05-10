# automodel-for-cc Windows installer
# Run: irm https://raw.githubusercontent.com/superunderpants/automodel-for-cc/master/install.ps1 | iex

$ErrorActionPreference = "Stop"
$repo = "superunderpants/automodel-for-cc"
$binName = "automodel-for-cc-windows-amd64.exe"
$installDir = "$env:APPDATA\auto-guard"
$binaryPath = "$installDir\automodel-for-cc.exe"
$configPath = "$installDir\config.yaml"
$settingsPath = "$env:USERPROFILE\.claude\settings.json"

Write-Host "=== automodel-for-cc installer ==="

# 1. Download binary
Write-Host "[1/3] Downloading binary..."
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
$url = "https://github.com/$repo/releases/latest/download/$binName"
Invoke-WebRequest -Uri $url -OutFile $binaryPath -UseBasicParsing
Write-Host "       -> $binaryPath"

# 2. LLM config — 3 fields, same as configuring Claude Code
Write-Host "[2/3] Configuring LLM review (Anthropic API)..."
Write-Host "       Press Enter to use the default shown in brackets."
Write-Host ""

$baseUrl = Read-Host "Base URL (use Anthropic endpoint)"
if (-not $baseUrl) { $baseUrl = "https://api.deepseek.com/anthropic" }

$apiKey = Read-Host "API Key"

$model = Read-Host "Model [deepseek-chat]"
if (-not $model) { $model = "deepseek-chat" }

# Write config.yaml
@"
# automodel-for-cc config — same 3 fields as Claude Code
llm:
  base_url: "$baseUrl"
  api_key: "$apiKey"
  model: "$model"
"@ | Set-Content $configPath -Encoding UTF8
Write-Host "       -> $configPath"

# 3. Hook
Write-Host "[3/3] Setting up Claude Code hook..."
$escapedPath = $binaryPath -replace '\\', '/'

if (Test-Path $settingsPath) {
    $settings = Get-Content $settingsPath -Raw | ConvertFrom-Json
} else {
    $settings = [pscustomobject]@{}
}

if (-not $settings.hooks) {
    $settings | Add-Member -MemberType NoteProperty -Name hooks -Value ([pscustomobject]@{})
}
if (-not $settings.hooks.PreToolUse) {
    $settings.hooks | Add-Member -MemberType NoteProperty -Name PreToolUse -Value @()
}

$newEntry = [pscustomobject]@{
    matcher = ".*"
    hooks   = @(
        [pscustomobject]@{
            type    = "command"
            command = $escapedPath
        }
    )
}

$found = $false
for ($i = 0; $i -lt $settings.hooks.PreToolUse.Count; $i++) {
    $cmd = $settings.hooks.PreToolUse[$i].hooks[0].command
    if ($cmd -and $cmd -match "automodel-for-cc") {
        $settings.hooks.PreToolUse[$i] = $newEntry
        $found = $true
        break
    }
}
if (-not $found) {
    $settings.hooks.PreToolUse += $newEntry
}

$settings | ConvertTo-Json -Depth 6 | Set-Content $settingsPath
Write-Host "       -> Hook added to $settingsPath"

Write-Host ""
Write-Host "Done! Restart Claude Code and you're all set."
Write-Host "Logs: $installDir\guard.log"
