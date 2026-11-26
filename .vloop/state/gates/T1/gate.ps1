$ErrorActionPreference = "Stop"
$root = Join-Path ([System.IO.Path]::GetTempPath()) ("vloop-b41-" + [Guid]::NewGuid().ToString("N"))
$repo = Join-Path $root "repo"
$exe = Join-Path $root "vloop.exe"
$oldPath = $env:PATH

try {
    & go build -o $exe ./cmd/vloop
    if ($LASTEXITCODE -ne 0) { throw "go build failed" }

    New-Item -ItemType Directory -Force -Path (Join-Path $repo ".vloop/state"), (Join-Path $repo "docs/briefs") | Out-Null
    & git init -q $repo
    if ($LASTEXITCODE -ne 0) { throw "git init failed" }
    & git -C $repo config user.name test
    & git -C $repo config user.email test@example.com
    & git -C $repo config commit.gpgsign false

    $briefA = "docs/briefs/B20260101-0900-a.loop-brief.md"
    $briefB = "docs/briefs/B20260101-0901-b.loop-brief.md"
    Set-Content (Join-Path $repo $briefA) "---`nname: B20260101-0900-a.loop-brief`nstatus: ready`n---`n"
    Set-Content (Join-Path $repo $briefB) "---`nname: B20260101-0901-b.loop-brief`nstatus: ready`n---`n"

    $pwshDir = Split-Path -Parent (Get-Command pwsh).Source
    $gitDir = Split-Path -Parent (Get-Command git).Source
    $separator = [System.IO.Path]::PathSeparator
    $pathWithConfigured = $pwshDir + $separator + $gitDir
    $pathWithoutConfigured = $gitDir

    function Write-Config([bool] $Mixed) {
        $body = 'shell = "pwsh"' + "`n"
        if ($Mixed) { $body += 'platforms = ["macos", "windows"]' + "`n" }
        Set-Content (Join-Path $repo ".vloop/config.toml") $body
    }

    function Write-Plan([string] $Schema, [string] $Status, [string] $Shell) {
        if ($Schema -eq "state/v1") {
            Set-Content (Join-Path $repo ".vloop/state/state.json") ('{"schema":"state/v1","status":"' + $Status + '","tasks":[]}' + "`n")
            return
        }
        $plan = @{
            schema = "state/v2"
            run_id = "B20260101-0900-a"
            brief = $briefA
            base = "0000000000000000000000000000000000000000"
            branch = "main"
            shell = $Shell
            status = $Status
            iteration = 0
            created = "2026-01-01T09:00:00Z"
            updated = "2026-01-01T09:00:00Z"
            checks = @()
            gate_scratch = @()
            gate_review = @{ rounds = 0; verdict = "" }
            tasks = @()
        }
        Set-Content (Join-Path $repo ".vloop/state/state.json") ($plan | ConvertTo-Json -Depth 5)
    }

    function Remove-Plan {
        Remove-Item (Join-Path $repo ".vloop/state/state.json") -Force -ErrorAction SilentlyContinue
    }

    function Doctor-Row([string] $PathValue) {
        $env:PATH = $PathValue
        $raw = & $exe -C $repo doctor --json 2>$null | Out-String
        if ([string]::IsNullOrWhiteSpace($raw)) { throw "doctor --json produced no output" }
        $doc = $raw | ConvertFrom-Json
        $row = $doc.checks | Where-Object { $_.name -eq "gate shell" } | Select-Object -First 1
        if ($null -eq $row) { throw "doctor --json omitted gate shell" }
        return $row
    }

    function Doctor-Text([string] $PathValue) {
        $env:PATH = $PathValue
        return (& $exe -C $repo doctor 2>&1 | Out-String)
    }

    function Assert-Doctor([string] $PathValue, [string] $Result, [string] $Message, [string] $Case) {
        $row = Doctor-Row $PathValue
        if ($row.result -ne $Result -or $row.message -ne $Message) {
            throw "$Case`: gate shell = $($row.result) '$($row.message)', want $Result '$Message'"
        }
    }

    function Run-Text([string] $PathValue, [string] $Brief) {
        $env:PATH = $PathValue
        if ($Brief) {
            return (& $exe -C $repo run $Brief 2>&1 | Out-String)
        }
        return (& $exe -C $repo run 2>&1 | Out-String)
    }

    function Commit-Fixture {
        & git -C $repo add -A
        & git -C $repo -c commit.gpgsign=false commit -q -m fixture
        if ($LASTEXITCODE -ne 0) { throw "git commit failed" }
    }

    $configuredMissingWarning = 'the configured shell "pwsh" is not on PATH'
    $configuredMissingProblem = $configuredMissingWarning
    $planMissingProblem = 'the plan''s shell "bash" is not on PATH'

    Remove-Plan
    Write-Config $true
    Assert-Doctor $pathWithConfigured "pass" "" "no plan, configured shell present"
    Write-Config $false
    Assert-Doctor $pathWithoutConfigured "warning" $configuredMissingWarning "no plan, configured shell absent, not mixed"
    Write-Config $true
    Assert-Doctor $pathWithoutConfigured "problem" $configuredMissingProblem "no plan, configured shell absent, mixed"

    Write-Plan "state/v2" "complete" "bash"
    Assert-Doctor $pathWithConfigured "pass" "" "complete state/v2 replaced, configured shell present"
    Assert-Doctor $pathWithoutConfigured "problem" $configuredMissingProblem "complete state/v2 replaced, configured shell absent, mixed"
    Write-Config $false
    Assert-Doctor $pathWithoutConfigured "warning" $configuredMissingWarning "complete state/v2 replaced, configured shell absent, not mixed"

    Write-Plan "state/v1" "complete" ""
    Write-Config $true
    Assert-Doctor $pathWithConfigured "pass" "" "complete state/v1, configured shell present"
    Assert-Doctor $pathWithoutConfigured "problem" $configuredMissingProblem "complete state/v1, configured shell absent, mixed"
    Write-Config $false
    Assert-Doctor $pathWithoutConfigured "warning" $configuredMissingWarning "complete state/v1, configured shell absent, not mixed"

    Write-Config $true
    foreach ($status in @("planning", "running", "blocked", "stalled", "halted")) {
        Write-Plan "state/v2" $status "bash"
        Assert-Doctor $pathWithConfigured "problem" $planMissingProblem "state/v2 $status keeps plan shell"
    }

    Write-Plan "state/v2" "complete" "bash"
    $workedPass = Doctor-Text $pathWithConfigured
    if ($workedPass -notmatch '✓ gate shell') {
        throw "worked example did not print the configured gate-shell pass: $workedPass"
    }
    $expectedPlanLine = "✓ plan a complete plan of " + $briefA + " — the next brief's plan replaces it"
    if ($workedPass -notmatch [Regex]::Escape($expectedPlanLine)) {
        throw "worked example did not print the replaced plan row: $workedPass"
    }

    Write-Plan "state/v2" "running" "bash"
    $workedPlanProblem = Doctor-Text $pathWithConfigured
    if ($workedPlanProblem -notmatch '✗ gate shell the plan''s shell "bash" is not on PATH') {
        throw "worked example did not print the in-use plan-shell problem: $workedPlanProblem"
    }

    Write-Plan "state/v2" "complete" "bash"
    Write-Config $false
    $workedWarning = Doctor-Text $pathWithoutConfigured
    if ($workedWarning -notmatch '! gate shell the configured shell "pwsh" is not on PATH') {
        throw "worked example did not print the configured-shell warning: $workedWarning"
    }
    Write-Config $true
    Commit-Fixture
    $different = Run-Text $pathWithConfigured $briefB
    if ($different -match "gate shell .*the plan's shell") {
        throw "preflight for another brief still judged the replaced plan's shell: $different"
    }

    Write-Config $false
    Commit-Fixture
    $differentWarning = Run-Text $pathWithoutConfigured $briefB
    if ($differentWarning -notmatch '! gate shell the configured shell "pwsh" is not on PATH') {
        throw "preflight for another brief did not carry the configured-shell warning: $differentWarning"
    }

    Write-Config $true
    Commit-Fixture
    $differentProblem = Run-Text $pathWithoutConfigured $briefB
    if ($differentProblem -notmatch '✗ gate shell the configured shell "pwsh" is not on PATH') {
        throw "preflight for another brief did not carry the configured-shell problem: $differentProblem"
    }

    foreach ($sameRun in @($briefA, "")) {
        $same = Run-Text $pathWithConfigured $sameRun
        if ($same -notmatch 'gate shell the plan''s shell "bash" is not on PATH') {
            throw "preflight for $sameRun did not judge the in-use plan's shell: $same"
        }
    }

    $doc = & $exe -C $repo doctor --json 2>$null | Out-String | ConvertFrom-Json
    $planRow = $doc.checks | Where-Object { $_.name -eq "plan" } | Select-Object -First 1
    $expectedPlan = "a complete plan of " + $briefA + " — the next brief's plan replaces it"
    if ($planRow.result -ne "pass" -or $planRow.message -ne $expectedPlan) {
        throw "worked example plan row = $($planRow.result) '$($planRow.message)', want pass '$expectedPlan'"
    }
}
finally {
    $env:PATH = $oldPath
    Remove-Item $root -Recurse -Force -ErrorAction SilentlyContinue
}
