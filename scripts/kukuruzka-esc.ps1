param(
    [Parameter(Position = 0)]
    [string]$Command = "help",

    [Parameter(Position = 1)]
    [string]$Arg1,

    [Parameter(Position = 2)]
    [string]$Arg2,

    [Parameter(Position = 3)]
    [string]$Arg3
)

$ErrorActionPreference = "Stop"

$Module = if ($env:KUKURUZKA_ESC_MODULE) { $env:KUKURUZKA_ESC_MODULE } else { "github.com/KOBAN13/kukuruzka-esc" }
$DefaultPath = if ($env:KUKURUZKA_ESC_PATH) { $env:KUKURUZKA_ESC_PATH } else { "..\kukuruzka-esc" }
$DefaultRef = if ($env:KUKURUZKA_ESC_REF) { $env:KUKURUZKA_ESC_REF } else { "latest" }
$DefaultInterval = if ($env:KUKURUZKA_ESC_INTERVAL) { $env:KUKURUZKA_ESC_INTERVAL } else { "10" }

function Show-Usage {
    Write-Host @"
Usage:
  scripts\kukuruzka-esc.cmd local [path]
      Use a local checkout of $Module through go.mod replace.

  scripts\kukuruzka-esc.cmd update [ref]
      Remove local replace, then update $Module from Git.

  scripts\kukuruzka-esc.cmd watch-update [ref] [seconds]
      Run update in a loop. The interval defaults to $DefaultInterval seconds.

  scripts\kukuruzka-esc.cmd publish [path] <commit-message> [ref]
      In the dependency repository: git add -A, commit if needed, push.
      Then update this project to the pushed ref.
"@
}

function Stop-WithError([string]$Message) {
    Write-Error $Message
    exit 1
}

function Assert-GoModRoot {
    if (-not (Test-Path "go.mod")) {
        Stop-WithError "run this script from the Velora repository root"
    }
}

function Assert-DependencyPath([string]$Path) {
    if (-not (Test-Path $Path -PathType Container)) {
        Stop-WithError "dependency path does not exist: $Path"
    }

    $goMod = Join-Path $Path "go.mod"
    if (-not (Test-Path $goMod -PathType Leaf)) {
        Stop-WithError "dependency path has no go.mod: $Path"
    }

    $hasModule = Select-String -Path $goMod -Pattern "^\s*module\s+$([regex]::Escape($Module))\s*$" -Quiet
    if (-not $hasModule) {
        Stop-WithError "dependency go.mod is not module $Module"
    }
}

function Invoke-Go {
    & go @args
    if ($LASTEXITCODE -ne 0) {
        throw "go command failed: go $args"
    }
}

function Invoke-Git {
    & git @args
    if ($LASTEXITCODE -ne 0) {
        throw "git command failed: git $args"
    }
}

function Show-ModuleState {
    Invoke-Go list -m -f "{{if .Replace}}{{.Path}} => {{.Replace.Path}}{{else}}{{.Path}} {{.Version}}{{end}}" $Module
}

function Use-Local([string]$Path) {
    if (-not $Path) {
        $Path = $DefaultPath
    }

    Assert-GoModRoot
    Assert-DependencyPath $Path

    $absolutePath = (Resolve-Path $Path).Path
    Invoke-Go mod edit "-replace=$Module=$absolutePath"
    Invoke-Go mod tidy
    Show-ModuleState
}

function Update-FromGit([string]$Ref) {
    if (-not $Ref) {
        $Ref = $DefaultRef
    }

    Assert-GoModRoot
    & go mod edit "-dropreplace=$Module" 2>$null
    Invoke-Go get "$Module@$Ref"
    Invoke-Go mod tidy
    Show-ModuleState
}

function Watch-UpdateFromGit([string]$Ref, [string]$Interval) {
    if (-not $Ref) {
        $Ref = $DefaultRef
    }
    if (-not $Interval) {
        $Interval = $DefaultInterval
    }
    if ($Interval -notmatch "^[1-9][0-9]*$") {
        Stop-WithError "interval must be a positive integer"
    }

    while ($true) {
        Write-Host "[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] updating $Module@$Ref"
        try {
            Update-FromGit $Ref
            Write-Host "[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] update finished"
        } catch {
            Write-Warning "update failed; retrying in $Interval seconds: $($_.Exception.Message)"
        }
        Start-Sleep -Seconds ([int]$Interval)
    }
}

function Publish-AndUpdate([string]$Path, [string]$Message, [string]$Ref) {
    if (-not $Path) {
        Stop-WithError "dependency path is required"
    }
    if (-not $Message) {
        Stop-WithError "commit message is required"
    }

    Assert-GoModRoot
    Assert-DependencyPath $Path

    Invoke-Git -C $Path add -A
    & git -C $Path diff --cached --quiet
    if ($LASTEXITCODE -ne 0) {
        Invoke-Git -C $Path commit -m $Message
    }

    Invoke-Git -C $Path push

    if (-not $Ref) {
        $Ref = (& git -C $Path branch --show-current).Trim()
        if (-not $Ref) {
            Stop-WithError "cannot infer ref from detached HEAD; pass tag, branch, or commit hash explicitly"
        }
    }

    Update-FromGit $Ref
}

switch ($Command) {
    "local" { Use-Local $Arg1 }
    "update" { Update-FromGit $Arg1 }
    "watch-update" { Watch-UpdateFromGit $Arg1 $Arg2 }
    "publish" { Publish-AndUpdate $Arg1 $Arg2 $Arg3 }
    "help" { Show-Usage }
    "-h" { Show-Usage }
    "--help" { Show-Usage }
    default {
        Show-Usage
        Stop-WithError "unknown command: $Command"
    }
}
