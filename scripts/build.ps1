#!/usr/bin/env pwsh
$ErrorActionPreference = 'Stop'

$mode = $args[0]
if (-not $mode) { $mode = 'build' }

# CGO is required (malgo audio + modernc sqlite). We default to a
# mingw-w64 compatible clang/gcc; install llvm-mingw if missing.
$env:CGO_ENABLED = '1'
$cc = Get-Command gcc -ErrorAction SilentlyContinue
if (-not $cc) { $cc = Get-Command clang -ErrorAction SilentlyContinue }
if (-not $cc) {
    Write-Error @"
No C compiler found on PATH. shushingface needs CGO to build.
Install one of:
  - llvm-mingw: https://github.com/mstorsjo/llvm-mingw/releases
    (pick the ucrt-aarch64 zip on ARM64 Windows, ucrt-x86_64 on x64)
  - MSYS2 mingw-w64-gcc
Then add its 'bin' directory to PATH and re-run this command.
"@
    exit 1
}

$version = (& git describe --tags --always --dirty 2>$null)
if ($LASTEXITCODE -ne 0 -or -not $version) { $version = 'dev' }
$ldflags = "-X codeberg.org/dbus/shushingface/internal/version.version=$version"

$wailsArgs = @($mode)
foreach ($a in $args[1..($args.Length - 1)]) { $wailsArgs += $a }
$wailsArgs += '-ldflags'
$wailsArgs += $ldflags

& wails @wailsArgs
exit $LASTEXITCODE
