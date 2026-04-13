#!/usr/bin/env pwsh
# Windows-idiomatic per-user install. Drops shushingface.exe into
# %LOCALAPPDATA%\Programs\shushingface\ (the convention for installs
# that don't require admin elevation) and creates a Start Menu shortcut.
# Cleans up earlier install paths if they exist.
$ErrorActionPreference = 'Stop'

$src = Join-Path $PSScriptRoot '..\..\build\bin\shushingface.exe'
if (-not (Test-Path $src)) {
    throw "build/bin/shushingface.exe not found - run 'just build' first"
}

$dest = Join-Path $env:LOCALAPPDATA 'Programs\shushingface'
$exe  = Join-Path $dest 'shushingface.exe'
$startMenu = Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs'
$lnk = Join-Path $startMenu 'shushingface.lnk'

# --- legacy-path cleanup ----------------------------------------------------
# An earlier refactor briefly installed under $HOME\.local\bin to mirror
# Linux PREFIX. Remove that too if we find it so users don't end up with
# two copies on PATH.
$legacyHomeBins = @()
if ($env:HOME)            { $legacyHomeBins += (Join-Path (Join-Path $env:HOME '.local') 'bin\shushingface.exe') }
if ($env:USERPROFILE)     { $legacyHomeBins += (Join-Path (Join-Path $env:USERPROFILE '.local') 'bin\shushingface.exe') }
foreach ($legacy in $legacyHomeBins) {
    if (Test-Path $legacy) {
        Write-Host "[warn] removing legacy install at $legacy" -ForegroundColor Yellow
        Remove-Item -Force $legacy
    }
}

# --- install ----------------------------------------------------------------
New-Item -ItemType Directory -Force -Path $dest | Out-Null
Copy-Item -Force $src $exe

New-Item -ItemType Directory -Force -Path $startMenu | Out-Null
$wsh = New-Object -ComObject WScript.Shell
$shortcut = $wsh.CreateShortcut($lnk)
$shortcut.TargetPath = $exe
$shortcut.WorkingDirectory = $dest
$shortcut.Save()

Write-Host "[info] installed shushingface to $exe" -ForegroundColor Blue
Write-Host "[info] launch it from the Start Menu (shushingface)" -ForegroundColor Blue
