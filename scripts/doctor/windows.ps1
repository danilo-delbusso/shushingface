#Requires -Version 5.1
# Windows-side dependency reporter. Mirrors scripts/doctor/linux.sh /
# common.sh but native PowerShell so it sees the same PATH the user's
# shell sees (Git Bash / WSL bash do not).
[CmdletBinding()]
param()
$ErrorActionPreference = 'Stop'

# Load tool versions shared with Linux + CI.
$versionsFile = Join-Path $PSScriptRoot '..\lib\versions.env'
$versions = @{}
Get-Content $versionsFile | ForEach-Object {
    if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
    if ($_ -match '^\s*([^=]+?)\s*=\s*(.+)\s*$') { $versions[$Matches[1]] = $Matches[2] }
}

$rows = New-Object System.Collections.Generic.List[object]

function Add-Row ($tool, $status, $version, $class, $action) {
    $rows.Add([pscustomobject]@{
        Tool = $tool; Status = $status; Version = $version; Class = $class; Action = $action
    })
}

# Note: parameter must NOT be named $args — that's a PowerShell automatic
# that shadows the declared parameter when used with @-splatting inside
# the function. Use $cmdArgs instead.
function Get-CmdVersion ($cmd, [string[]]$cmdArgs) {
    try { (& $cmd @cmdArgs 2>$null) -join "`n" } catch { $null }
}

function Check ($tool, $cmd, [string[]]$verArgs, [string]$class, [string]$action, [scriptblock]$parser = $null) {
    $found = Get-Command $cmd -ErrorAction SilentlyContinue
    if (-not $found) {
        Add-Row $tool 'missing' '' $class $action
        return
    }
    $raw = Get-CmdVersion $cmd $verArgs
    $ver = if ($parser) { & $parser $raw } else { $raw -replace '\r','' }
    if ($ver -is [array]) { $ver = $ver[0] }
    if ([string]::IsNullOrWhiteSpace($ver)) { $ver = 'installed' }
    Add-Row $tool 'ok' $ver.Trim() $class ''
}

Write-Host "==> checking build / runtime dependencies (windows)" -ForegroundColor White

Check 'go'   'go'   @('env','GOVERSION') 'required' "install go $($versions['GO_VERSION'])+" `
    { param($v) ($v -replace '^go','') }

Check 'bun'  'bun'  @('--version') 'required' "install bun $($versions['BUN_VERSION'])+"

Check 'wails' 'wails' @('version') 'required' "go install github.com/wailsapp/wails/v2/cmd/wails@$($versions['WAILS_VERSION'])" `
    { param($v) (($v -split "`n") | Select-String '\d+\.\d+\.\d+' -List | ForEach-Object { $_.Matches[0].Value }) }

Check 'git'  'git'  @('--version') 'required' 'install Git for Windows' `
    { param($v) ($v -split ' ')[2] }

Check 'just' 'just' @('--version') 'required' 'winget install --exact --id Casey.Just' `
    { param($v) ($v -split ' ')[1] }

# CGO compiler
$cgo = Get-Command gcc -ErrorAction SilentlyContinue
if (-not $cgo) { $cgo = Get-Command clang -ErrorAction SilentlyContinue }
if ($cgo) {
    $ver = (& $cgo.Source -dumpversion 2>$null)
    Add-Row $cgo.Name 'ok' $ver 'required' ''
} else {
    Add-Row 'cgo-compiler' 'missing' '' 'required' 'install llvm-mingw (winget install -e --id mstorsjo.LLVM-MinGW.UCRT)'
}

# winget
if (Get-Command winget -ErrorAction SilentlyContinue) {
    Add-Row 'winget' 'ok' '' 'recommended' ''
} else {
    Add-Row 'winget' 'missing' '' 'recommended' 'install App Installer from Microsoft Store'
}

# NSIS / makensis (needed by `just package`). winget installs to
# Program Files but does NOT update PATH, so probe both.
$nsisPath = $null
foreach ($cand in @('C:\Program Files (x86)\NSIS\makensis.exe', 'C:\Program Files\NSIS\makensis.exe')) {
    if (Test-Path $cand) { $nsisPath = $cand; break }
}
if (Get-Command makensis -ErrorAction SilentlyContinue) {
    $nver = (& makensis -VERSION 2>$null) -replace '^v',''
    Add-Row 'makensis' 'ok' $nver 'recommended' ''
} elseif ($nsisPath) {
    $nver = (& $nsisPath -VERSION 2>$null) -replace '^v',''
    Add-Row 'makensis' 'ok' $nver 'recommended' "found at $($nsisPath) (auto-PATHed by 'just package')"
} else {
    Add-Row 'makensis' 'missing' '' 'recommended' "needed by 'just package' (winget install NSIS.NSIS)"
}

# golangci-lint (optional)
if (Get-Command golangci-lint -ErrorAction SilentlyContinue) {
    $glver = (golangci-lint version --short 2>$null)
    Add-Row 'golangci-lint' 'ok' $glver 'optional' ''
} else {
    Add-Row 'golangci-lint' 'missing' '' 'optional' 'optional: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest'
}

# --- render -----------------------------------------------------------------
Write-Host ""
$fmt = "{0,-16}  {1,-12}  {2,-14}  {3,-12}  {4}"
Write-Host ($fmt -f 'Tool','Status','Version','Class','Action')
Write-Host ('-' * 70)
$exit = 0
foreach ($r in $rows) {
    $mark = switch ($r.Status) {
        'ok'       { 'ok' }
        'missing'  { 'X missing' }
        default    { $r.Status }
    }
    $ver = if ([string]::IsNullOrWhiteSpace($r.Version)) { '-' } else { $r.Version }
    Write-Host ($fmt -f $r.Tool, $mark, $ver, $r.Class, $r.Action)
    if ($r.Status -ne 'ok' -and $r.Class -eq 'required') { $exit = 1 }
}
Write-Host ""
if ($exit -eq 0) {
    Write-Host "[info] all required dependencies present" -ForegroundColor Blue
} else {
    Write-Host "[warn] missing required dependencies - run 'just bootstrap' to install" -ForegroundColor Yellow
}
exit $exit
