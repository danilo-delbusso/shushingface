#!/usr/bin/env pwsh
# Windows build dispatcher. First arg is the mode: `build`, `dev`, or a
# literal wails subcommand. Remaining args are forwarded to `wails`.
#
# Mirrors scripts/build/linux.sh contract: sources VERSION / LDFLAGS
# from the same conventions (git describe, fallback "dev"), and stamps
# wails.json's info.productVersion to a numeric MAJOR.MINOR.PATCH pulled
# from the nearest git tag so NSIS VIProductVersion / Add-Remove-Programs
# metadata actually track the release.
$ErrorActionPreference = 'Stop'

$mode = $args[0]
if (-not $mode) { $mode = 'build' }

# CGO is required (malgo audio + modernc sqlite). Accept gcc or clang.
$env:CGO_ENABLED = '1'
$cc = Get-Command gcc -ErrorAction SilentlyContinue
if (-not $cc) { $cc = Get-Command clang -ErrorAction SilentlyContinue }
if (-not $cc) {
    Write-Error @"
No C compiler found on PATH. shushingface needs CGO to build.
Run `just bootstrap` to install llvm-mingw, or install one of:
  - llvm-mingw: https://github.com/mstorsjo/llvm-mingw/releases
    (pick the ucrt-aarch64 zip on ARM64 Windows, ucrt-x86_64 on x64)
  - MSYS2 mingw-w64-gcc
Then add its 'bin' directory to PATH and re-run this command.
"@
    exit 1
}

$version = (& git describe --tags --always --dirty --exclude 'backup/*' 2>$null)
if ($LASTEXITCODE -ne 0 -or -not $version) { $version = 'dev' }
$ldflags = "-X codeberg.org/dbus/shushingface/internal/version.version=$version"

# Numeric MAJOR.MINOR.PATCH for wails.json (NSIS VIProductVersion needs numerics).
$tag = (& git describe --tags --abbrev=0 --exclude 'backup/*' 2>$null)
if ($LASTEXITCODE -ne 0 -or -not $tag) { $tag = '' }
$tag = $tag -replace '^v', ''
if ($tag -match '^(\d+\.\d+\.\d+)') {
    $productVersion = $matches[1]
} else {
    $productVersion = '0.0.0'
}

# Stamp wails.json, snapshot original so we can restore in `finally`.
$wailsJson = 'wails.json'
$wailsBackup = $null
if (Test-Path $wailsJson) {
    $wailsBackup = Get-Content $wailsJson -Raw
    $replacement = '${1}' + $productVersion + '${2}'
    $patched = $wailsBackup -replace '("productVersion"\s*:\s*")[^"]*(")', $replacement
    if ($patched -ne $wailsBackup) {
        [System.IO.File]::WriteAllText((Resolve-Path $wailsJson), $patched, (New-Object System.Text.UTF8Encoding $false))
    }
}

try {
    # `wails dev` and `wails build -debug` hardcode `-gcflags=all=-N -l`,
    # which on windows/arm64 + Go 1.26 trips `syscall.Syscall15: nosplit stack
    # over 792 byte limit`. Until that lands upstream, dev mode falls back to
    # an optimised build (no -devtools — we don't want the right-click menu
    # by default) and re-launches the binary. Frontend changes require a
    # rebuild; there's no hot reload on this target.
    if ($mode -eq 'dev') {
        & wails build -ldflags $ldflags
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
        & ./build/bin/shushingface.exe
        exit $LASTEXITCODE
    }

    $wailsArgs = @($mode)
    if ($args.Length -gt 1) {
        foreach ($a in $args[1..($args.Length - 1)]) { $wailsArgs += $a }
    }
    $wailsArgs += '-ldflags'
    $wailsArgs += $ldflags

    & wails @wailsArgs
    exit $LASTEXITCODE
}
finally {
    if ($null -ne $wailsBackup) {
        [System.IO.File]::WriteAllText((Resolve-Path $wailsJson), $wailsBackup, (New-Object System.Text.UTF8Encoding $false))
    }
}
