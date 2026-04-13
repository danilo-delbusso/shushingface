#Requires -Version 5.1
# Install missing Windows dev dependencies via winget + go install.
# Accepts -Yes to skip confirmation. Respects $env:DRY_RUN=1.
[CmdletBinding()]
param(
    [switch]$Yes
)
$ErrorActionPreference = 'Stop'

$dryRun = ($env:DRY_RUN -eq '1')

# Read the tool-version manifest shared with Linux / CI.
$versionsFile = Join-Path $PSScriptRoot '..\lib\versions.env'
if (-not (Test-Path $versionsFile)) { throw "missing $versionsFile" }
$versions = @{}
Get-Content $versionsFile | ForEach-Object {
    if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
    if ($_ -match '^\s*([^=]+?)\s*=\s*(.+)\s*$') { $versions[$Matches[1]] = $Matches[2] }
}

function Info ($msg)  { Write-Host "[info] $msg" -ForegroundColor Blue }
function Warn ($msg)  { Write-Host "[warn] $msg" -ForegroundColor Yellow }
function Section ($msg) { Write-Host "`n==> $msg" -ForegroundColor White }

function Invoke-Install ([string]$Description, [string]$Cmd) {
    Section $Description
    Write-Host "  > $Cmd"
    if ($dryRun) { Write-Host "  (dry run - not executing)" -ForegroundColor DarkGray; return }
    if (-not $Yes) {
        $ans = Read-Host "  proceed? [y/N]"
        if ($ans -notmatch '^(y|Y|yes|YES)$') { Warn "skipped"; return }
    }
    Invoke-Expression $Cmd
}

if (-not (Get-Command winget -ErrorAction SilentlyContinue)) {
    throw "winget not available. Install 'App Installer' from the Microsoft Store first."
}

# --- system toolchain --------------------------------------------------------

if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    Invoke-Install "Git for Windows (provides Git Bash + git)" `
        "winget install --exact --id Git.Git --accept-source-agreements --accept-package-agreements"
}

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Invoke-Install "Go $($versions['GO_VERSION'])" `
        "winget install --exact --id GoLang.Go --accept-source-agreements --accept-package-agreements"
}

if (-not (Get-Command bun -ErrorAction SilentlyContinue)) {
    Invoke-Install "Bun $($versions['BUN_VERSION'])" `
        "winget install --exact --id Oven-sh.Bun --accept-source-agreements --accept-package-agreements"
}

if (-not ((Get-Command gcc -ErrorAction SilentlyContinue) -or (Get-Command clang -ErrorAction SilentlyContinue))) {
    Invoke-Install "LLVM-MinGW (CGO compiler)" `
        "winget install --exact --id mstorsjo.LLVM-MinGW.UCRT --accept-source-agreements --accept-package-agreements"
}

if (-not (Get-Command just -ErrorAction SilentlyContinue)) {
    Invoke-Install "just (task runner)" `
        "winget install --exact --id Casey.Just --accept-source-agreements --accept-package-agreements"
}

# NSIS provides makensis, used by `just package` to build the installer.
$nsisInstalled = (Get-Command makensis -ErrorAction SilentlyContinue) `
    -or (Test-Path 'C:\Program Files (x86)\NSIS\makensis.exe') `
    -or (Test-Path 'C:\Program Files\NSIS\makensis.exe')
if (-not $nsisInstalled) {
    Invoke-Install "NSIS (installer builder for 'just package')" `
        "winget install --exact --id NSIS.NSIS --accept-source-agreements --accept-package-agreements"
}

# --- Go-installed tools ------------------------------------------------------

if (-not (Get-Command wails -ErrorAction SilentlyContinue)) {
    Invoke-Install "wails CLI $($versions['WAILS_VERSION'])" `
        "go install github.com/wailsapp/wails/v2/cmd/wails@$($versions['WAILS_VERSION'])"
}

Info "bootstrap complete - run 'just doctor' to verify. You may need to restart your shell so new PATH entries are picked up."
