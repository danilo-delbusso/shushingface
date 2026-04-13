#!/usr/bin/env pwsh
$ErrorActionPreference = 'Stop'

$mode = $args[0]
if (-not $mode) { $mode = 'build' }

$version = (git describe --tags --always --dirty 2>$null)
if (-not $version) { $version = 'dev' }
$ldflags = "-X codeberg.org/dbus/shushingface/internal/version.version=$version"

$wailsArgs = @($mode)
foreach ($a in $args[1..($args.Length - 1)]) { $wailsArgs += $a }
$wailsArgs += '-ldflags'
$wailsArgs += $ldflags

& wails @wailsArgs
exit $LASTEXITCODE
