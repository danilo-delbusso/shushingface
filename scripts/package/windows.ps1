#!/usr/bin/env pwsh
# Build and package a Windows release: NSIS installer + standalone exe.
# Delegates the actual build to scripts/build/windows.ps1 so the CGO
# compiler check and nosplit-stack workaround live in one place.
$ErrorActionPreference = 'Stop'

$root = Resolve-Path (Join-Path $PSScriptRoot '..\..')
$build = Join-Path $PSScriptRoot '..\build\windows.ps1'

Write-Host "[info] packaging shushingface for windows" -ForegroundColor Blue

# Wails shells out to `makensis` for the installer step. winget's NSIS
# package installs to Program Files but does NOT add it to PATH, so add
# it here for the duration of the build if we find it.
if (-not (Get-Command makensis -ErrorAction SilentlyContinue)) {
    foreach ($cand in @('C:\Program Files (x86)\NSIS', 'C:\Program Files\NSIS')) {
        if (Test-Path (Join-Path $cand 'makensis.exe')) {
            $env:PATH = "$cand;$env:PATH"
            Write-Host "[info] using NSIS from $cand" -ForegroundColor Blue
            break
        }
    }
    if (-not (Get-Command makensis -ErrorAction SilentlyContinue)) {
        Write-Host "[warn] makensis not found - installer step will be skipped (run 'just bootstrap')" -ForegroundColor Yellow
    }
}

# Clean previous build
$binDir = Join-Path $root 'build\bin'
if (Test-Path $binDir) { Remove-Item -Recurse -Force $binDir }

# Build with NSIS installer (wails handles the NSIS script internally).
& powershell.exe -NoProfile -ExecutionPolicy Bypass -File $build build -nsis
if ($LASTEXITCODE -ne 0) { throw "wails build failed" }

$distDir = Join-Path $root 'dist'
New-Item -ItemType Directory -Force -Path $distDir | Out-Null

# Copy everything wails produced in build/bin into dist/.
Get-ChildItem -Path $binDir | ForEach-Object {
    Copy-Item -Force $_.FullName -Destination $distDir
    Write-Host "[info] packaged $($_.Name)" -ForegroundColor Blue
}
