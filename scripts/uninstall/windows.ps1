#!/usr/bin/env pwsh
# Removes shushingface from $PREFIX\bin and the Start Menu. Also cleans
# the legacy %LOCALAPPDATA%\Programs\shushingface install path.
$ErrorActionPreference = 'Stop'

$prefix = $env:PREFIX
if (-not $prefix) {
    if ($env:HOME)            { $prefix = Join-Path $env:HOME '.local' }
    elseif ($env:USERPROFILE) { $prefix = Join-Path $env:USERPROFILE '.local' }
}

$removed = @()

$newExe = Join-Path (Join-Path $prefix 'bin') 'shushingface.exe'
if ($prefix -and (Test-Path $newExe)) {
    Remove-Item -Force $newExe
    $removed += $newExe
}

$legacyDir = Join-Path $env:LOCALAPPDATA 'Programs\shushingface'
$legacyExe = Join-Path $legacyDir 'shushingface.exe'
if (Test-Path $legacyExe) {
    Remove-Item -Force $legacyExe
    $removed += $legacyExe
    if ((Get-ChildItem -Path $legacyDir -ErrorAction SilentlyContinue | Measure-Object).Count -eq 0) {
        Remove-Item -Force $legacyDir
    }
}

$lnk = Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs\shushingface.lnk'
if (Test-Path $lnk) { Remove-Item -Force $lnk; $removed += $lnk }

if ($removed.Count -eq 0) {
    Write-Host "[info] nothing to uninstall" -ForegroundColor Blue
} else {
    foreach ($r in $removed) { Write-Host "[info] removed $r" -ForegroundColor Blue }
}
