$ErrorActionPreference = "Stop"
$script:RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..\..")).Path
$script:RealGo = (Get-Command go -ErrorAction Stop).Source
$script:OriginalPath = $env:Path

function Assert-Gate {
    param([bool]$Condition, [string]$Message)
    if (-not $Condition) { throw $Message }
}

function Assert-Equal {
    param($Actual, $Expected, [string]$Message)
    if ($Actual -cne $Expected) {
        throw "$Message`nexpected: <$Expected>`nactual:   <$Actual>"
    }
}

function Format-GatePath {
    param([string]$Path)
    $full = [IO.Path]::GetFullPath($Path).Replace('\', '/')
    $home = [IO.Path]::GetFullPath([Environment]::GetFolderPath('UserProfile')).TrimEnd('\', '/').Replace('\', '/')
    if ($full.Equals($home, [StringComparison]::OrdinalIgnoreCase)) { return "~" }
    if ($full.StartsWith($home + '/', [StringComparison]::OrdinalIgnoreCase)) {
        return "~" + $full.Substring($home.Length)
    }
    return $full
}

function New-GateHarness {
    param(
        [string]$Module = "v0.14.0",
        [string]$Version = "0.14.0",
        [string]$Latest = "v0.15.0",
        [switch]$ExeOutsideGobin,
        [string[]]$Agents = @()
    )
    $temp = Join-Path ([IO.Path]::GetTempPath()) ("vloop-gate-" + [Guid]::NewGuid().ToString("N"))
    $bin = Join-Path $temp "bin"
    $fake = Join-Path $temp "fake"
    $gobin = Join-Path $temp "gobin"
    $run = Join-Path $temp "run"
    New-Item -ItemType Directory -Force $fake, $gobin, $run | Out-Null
    $goName = if ([OperatingSystem]::IsWindows()) { "go.exe" } else { "go" }
    & $script:RealGo build -o (Join-Path $fake $goName) (Join-Path $PSScriptRoot "fakego.go")
    if ($LASTEXITCODE -ne 0) { throw "building fake go failed" }
    & $script:RealGo build -o (Join-Path $fake "fakeagent.exe") (Join-Path $PSScriptRoot "fakeagent.go")
    if ($LASTEXITCODE -ne 0) { throw "building fake agent failed" }
    foreach ($agent in $Agents) {
        $agentName = if ([OperatingSystem]::IsWindows()) { $agent + ".exe" } else { $agent }
        Copy-Item (Join-Path $fake "fakeagent.exe") (Join-Path $fake $agentName)
    }
    $exeDir = if ($ExeOutsideGobin) { $run } else { $gobin }
    $exeName = if ([OperatingSystem]::IsWindows()) { "vloop.exe" } else { "vloop" }
    $exe = Join-Path $exeDir $exeName
    New-Item -ItemType Directory -Force $exeDir | Out-Null
    & $script:RealGo build -o $exe (Join-Path $PSScriptRoot "runner.go")
    if ($LASTEXITCODE -ne 0) { throw "building vloop gate runner failed" }

    $env:VLOOP_GATE_MODULE = $Module
    $env:VLOOP_GATE_VERSION = $Version
    $env:FAKE_GOBIN = $gobin
    $env:FAKE_GOPATH = ""
    $env:FAKE_GO_LATEST = $Latest
    $env:FAKE_GO_LIST_FAIL = ""
    $env:FAKE_GO_ENV_FAIL = ""
    $env:FAKE_GO_INSTALL_FAIL = ""
    $env:FAKE_GO_LOG = Join-Path $temp "go.log"
    $env:FAKE_AGENT_LOG = Join-Path $temp "agent.log"
    $env:FAKE_CLAUDE_LIST = "[]"
    $env:FAKE_COPILOT_LIST = "[]"
    $env:FAKE_CLAUDE_LIST_FAIL = ""
    $env:FAKE_COPILOT_LIST_FAIL = ""
    $env:FAKE_CLAUDE_MARKETPLACE_FAIL = ""
    $env:FAKE_CLAUDE_UPDATE_FAIL = ""
    $env:FAKE_COPILOT_UPDATE_FAIL = ""
    $env:Path = $fake

    return [pscustomobject]@{
        Temp = $temp
        Bin = $bin
        Fake = $fake
        Gobin = $gobin
        Exe = $exe
        GoLog = $env:FAKE_GO_LOG
        AgentLog = $env:FAKE_AGENT_LOG
        ExeName = $exeName
    }
}

function Invoke-GateCommand {
    param([string]$Exe, [string[]]$Arguments, [string]$WorkDir = $script:RepoRoot)
    $id = [Guid]::NewGuid().ToString("N")
    $stdout = Join-Path ([IO.Path]::GetTempPath()) ("vloop-gate-out-" + $id)
    $stderr = Join-Path ([IO.Path]::GetTempPath()) ("vloop-gate-err-" + $id)
    Push-Location $WorkDir
    try {
        & $Exe @Arguments 1> $stdout 2> $stderr
        $code = $LASTEXITCODE
    } finally {
        Pop-Location
    }
    $out = if (Test-Path $stdout) { [IO.File]::ReadAllText($stdout).TrimEnd("`r", "`n") } else { "" }
    $err = if (Test-Path $stderr) { [IO.File]::ReadAllText($stderr).TrimEnd("`r", "`n") } else { "" }
    Remove-Item $stdout, $stderr -Force -ErrorAction SilentlyContinue
    return [pscustomobject]@{ Code = $code; Out = $out; Err = $err }
}

function Get-GateLog {
    param([string]$Path)
    if (-not (Test-Path $Path)) { return "" }
    return [IO.File]::ReadAllText($Path).TrimEnd("`r", "`n")
}
