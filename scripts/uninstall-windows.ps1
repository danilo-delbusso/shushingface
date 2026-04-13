#!/usr/bin/env pwsh
$ErrorActionPreference = 'SilentlyContinue'

$dest = Join-Path $env:LOCALAPPDATA 'Programs\shushingface'
Remove-Item -Force (Join-Path $dest 'shushingface.exe')
Remove-Item -Force $dest
Remove-Item -Force (Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs\shushingface.lnk')

Write-Host "Uninstalled shushingface"
