#!/usr/bin/env pwsh
# Installs shushingface.exe under $PREFIX\bin (default
# %USERPROFILE%\.local\bin) and creates a Start Menu shortcut.
# Migrates and cleans up the legacy %LOCALAPPDATA%\Programs\shushingface
# install path used before the PREFIX unification.
$ErrorActionPreference = 'Stop'

$src = Join-Path $PSScriptRoot '..\..\build\bin\shushingface.exe'
if (-not (Test-Path $src)) {
    throw "build/bin/shushingface.exe not found - run 'just build' first"
}

# Resolve PREFIX with the same fallback chain as scripts/lib/version.sh:
# $PREFIX -> $HOME\.local -> $USERPROFILE\.local.
$prefix = $env:PREFIX
if (-not $prefix) {
    if ($env:HOME)              { $prefix = Join-Path $env:HOME '.local' }
    elseif ($env:USERPROFILE)   { $prefix = Join-Path $env:USERPROFILE '.local' }
    else { throw "cannot resolve PREFIX (set \$env:PREFIX or HOME)" }
}

$binDir = Join-Path $prefix 'bin'
$dest = Join-Path $binDir 'shushingface.exe'
$startMenu = Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs'
$lnk = Join-Path $startMenu 'shushingface.lnk'

# --- legacy-path migration ---------------------------------------------------
# Old install used %LOCALAPPDATA%\Programs\shushingface\shushingface.exe.
# If found, remove it and its old Start Menu shortcut before installing.
$legacyDir = Join-Path $env:LOCALAPPDATA 'Programs\shushingface'
$legacyExe = Join-Path $legacyDir 'shushingface.exe'
if (Test-Path $legacyExe) {
    Write-Host "[warn] migrating install from legacy path $legacyDir" -ForegroundColor Yellow
    Remove-Item -Force $legacyExe
    if ((Get-ChildItem -Path $legacyDir -ErrorAction SilentlyContinue | Measure-Object).Count -eq 0) {
        Remove-Item -Force $legacyDir
    }
}

# --- install ----------------------------------------------------------------
New-Item -ItemType Directory -Force -Path $binDir | Out-Null
Copy-Item -Force $src $dest

New-Item -ItemType Directory -Force -Path $startMenu | Out-Null
$wsh = New-Object -ComObject WScript.Shell
$shortcut = $wsh.CreateShortcut($lnk)
$shortcut.TargetPath = $dest
$shortcut.WorkingDirectory = $binDir
$shortcut.Save()

Write-Host "[info] installed shushingface to $dest" -ForegroundColor Blue
Write-Host "[info] tip: ensure $binDir is on your PATH" -ForegroundColor Blue
