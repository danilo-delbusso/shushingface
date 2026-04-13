#!/usr/bin/env pwsh
# Installs shushingface.exe into %LOCALAPPDATA%\Programs\shushingface
# and creates a Start Menu shortcut.
$ErrorActionPreference = 'Stop'

$src = Join-Path $PSScriptRoot '..\build\bin\shushingface.exe'
if (-not (Test-Path $src)) {
    throw "build/bin/shushingface.exe not found - run 'just build-windows' first"
}

$dest = Join-Path $env:LOCALAPPDATA 'Programs\shushingface'
New-Item -ItemType Directory -Force -Path $dest | Out-Null
Copy-Item -Force $src (Join-Path $dest 'shushingface.exe')

$startMenu = Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs'
$lnk = Join-Path $startMenu 'shushingface.lnk'
$wsh = New-Object -ComObject WScript.Shell
$shortcut = $wsh.CreateShortcut($lnk)
$shortcut.TargetPath = Join-Path $dest 'shushingface.exe'
$shortcut.WorkingDirectory = $dest
$shortcut.Save()

Write-Host "Installed shushingface to $dest"
Write-Host "Tip: open the app and bind a shortcut from Settings -> Shortcut"
