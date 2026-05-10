# automodel-for-cc Windows installer
# Run: irm https://raw.githubusercontent.com/superunderpants/automodel-for-cc/master/install.ps1 | iex
param([string]$ApiKey)

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

# 2. API key
Write-Host "[2/3] Configuring API key..."
if (-not $ApiKey) {
    $ApiKey = [Environment]::GetEnvironmentVariable("DEEPSEEK_API_KEY", "User")
}
if (-not $ApiKey) {
    $ApiKey = Read-Host "Enter your DeepSeek API key (or press Enter to skip)"
}
if ($ApiKey) {
    [Environment]::SetEnvironmentVariable("DEEPSEEK_API_KEY", $ApiKey, "User")
    Write-Host "       -> DEEPSEEK_API_KEY set (restart terminal to take effect)"
} else {
    Write-Host "       -> skipped (set DEEPSEEK_API_KEY manually)"
}

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
