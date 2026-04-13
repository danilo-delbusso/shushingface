#!/usr/bin/env pwsh
# Windows build dispatcher. First arg is the mode: `build`, `dev`, or a
# literal wails subcommand. Remaining args are forwarded to `wails`.
#
# Mirrors scripts/build/linux.sh contract: sources VERSION / LDFLAGS
# from the same conventions (git describe, fallback "dev").
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

# Real dev mode: vite hot reload + debug build. Earlier wails / Go combos
# tripped a nosplit-stack limit on windows/arm64 because of
# -gcflags=all=-N -l, but as of wails v2.12 + Go 1.26.x this builds. If
# you regress here, fall back to `wails build -ldflags` and re-launch the
# binary manually.
if ($mode -eq 'dev') {
    & wails dev -ldflags $ldflags
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
