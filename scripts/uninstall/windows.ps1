#!/usr/bin/env pwsh
# Removes shushingface from %LOCALAPPDATA%\Programs\shushingface and the
# Start Menu. Also cleans the short-lived $HOME\.local\bin install path
# from the PREFIX-unification experiment.
$ErrorActionPreference = 'Stop'

$removed = @()

$dest = Join-Path $env:LOCALAPPDATA 'Programs\shushingface'
$exe  = Join-Path $dest 'shushingface.exe'
if (Test-Path $exe) {
    Remove-Item -Force $exe
    $removed += $exe
    if ((Get-ChildItem -Path $dest -ErrorAction SilentlyContinue | Measure-Object).Count -eq 0) {
        Remove-Item -Force $dest
    }
}

$legacyHomeBins = @()
if ($env:HOME)        { $legacyHomeBins += (Join-Path (Join-Path $env:HOME '.local') 'bin\shushingface.exe') }
if ($env:USERPROFILE) { $legacyHomeBins += (Join-Path (Join-Path $env:USERPROFILE '.local') 'bin\shushingface.exe') }
foreach ($legacy in $legacyHomeBins) {
    if (Test-Path $legacy) { Remove-Item -Force $legacy; $removed += $legacy }
}

$lnk = Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs\shushingface.lnk'
if (Test-Path $lnk) { Remove-Item -Force $lnk; $removed += $lnk }

if ($removed.Count -eq 0) {
    Write-Host "[info] nothing to uninstall" -ForegroundColor Blue
} else {
    foreach ($r in $removed) { Write-Host "[info] removed $r" -ForegroundColor Blue }
}
