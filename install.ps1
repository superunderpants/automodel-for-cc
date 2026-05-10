# automodel-for-cc Windows installer
# Run: irm https://raw.githubusercontent.com/superunderpants/automodel-for-cc/master/install.ps1 | iex

$ErrorActionPreference = "Stop"
$repo = "superunderpants/automodel-for-cc"
$binName = "automodel-for-cc-windows-amd64.exe"
$installDir = "$env:APPDATA\auto-guard"
$binaryPath = "$installDir\automodel-for-cc.exe"
$settingsPath = "$env:USERPROFILE\.claude\settings.json"

Write-Host "=== automodel-for-cc installer ==="

# 1. Download binary
Write-Host "[1/3] Downloading binary..."
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
$url = "https://github.com/$repo/releases/latest/download/$binName"
Invoke-WebRequest -Uri $url -OutFile $binaryPath -UseBasicParsing
Write-Host "       -> $binaryPath"

# 2. Provider & API key
Write-Host "[2/3] Configuring LLM provider..."
Write-Host ""
Write-Host "  [1] DeepSeek"
Write-Host "  [2] OpenAI"
Write-Host "  [3] Anthropic"
Write-Host "  [4] OpenRouter"
Write-Host "  [5] Ollama (local, no API key needed)"
Write-Host "  [6] Custom (any OpenAI-compatible API)"
Write-Host ""

$providers = @{
    "1" = @{name="deepseek";  env="DEEPSEEK_API_KEY"}
    "2" = @{name="openai";    env="OPENAI_API_KEY"}
    "3" = @{name="anthropic"; env="ANTHROPIC_API_KEY"}
    "4" = @{name="openrouter"; env="OPENROUTER_API_KEY"}
    "5" = @{name="ollama";    env=""}
    "6" = @{name="custom";    env="AUTO_GUARD_API_KEY"}
}

$choice = Read-Host "Choose provider [1-6]"
if (-not $providers.ContainsKey($choice) -or $choice -eq "0") {
    Write-Host "       -> Skipping, defaulting to DeepSeek"
    $provider = "deepseek"
    $envKey = "DEEPSEEK_API_KEY"
} else {
    $provider = $providers[$choice].name
    $envKey = $providers[$choice].env
}

Write-Host "       -> Provider: $provider"

if ($choice -eq "5") {
    Write-Host "       -> Ollama doesn't need an API key (local)"
} elseif ($choice -eq "6") {
    $baseUrl = Read-Host "Enter base URL"
    [Environment]::SetEnvironmentVariable("AUTO_GUARD_BASE_URL", $baseUrl, "User")
    $apiKey = Read-Host "Enter API key (or press Enter to skip)"
    if ($apiKey) {
        [Environment]::SetEnvironmentVariable("AUTO_GUARD_API_KEY", $apiKey, "User")
    }
} else {
    $apiKey = Read-Host "Enter API key (or press Enter to skip)"
    if ($apiKey) {
        [Environment]::SetEnvironmentVariable($envKey, $apiKey, "User")
    }
}

if ($provider -ne "deepseek") {
    [Environment]::SetEnvironmentVariable("AUTO_GUARD_PROVIDER", $provider, "User")
}

Write-Host "       -> Done (restart terminal to take effect)"

# 3. Hook
Write-Host "[3/3] Setting up Claude Code hook..."
$escapedPath = $binaryPath -replace '\\', '/'

# Read or create settings
if (Test-Path $settingsPath) {
    $settings = Get-Content $settingsPath -Raw | ConvertFrom-Json
} else {
    $settings = [pscustomobject]@{}
}

# Ensure hooks.PreToolUse exists
if (-not $settings.hooks) {
    $settings | Add-Member -MemberType NoteProperty -Name hooks -Value ([pscustomobject]@{})
}
if (-not $settings.hooks.PreToolUse) {
    $settings.hooks | Add-Member -MemberType NoteProperty -Name PreToolUse -Value @()
}

# Replace existing auto-guard entry or add new one
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

# Write back
$settings | ConvertTo-Json -Depth 6 | Set-Content $settingsPath
Write-Host "       -> Hook added to $settingsPath"

Write-Host ""
Write-Host "Done! Restart Claude Code and you're all set."
Write-Host "To verify: type anything in Claude Code — you should see fewer permission prompts."
