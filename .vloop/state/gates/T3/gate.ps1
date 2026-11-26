$ErrorActionPreference = 'Stop'

& pwsh -NoProfile -File .vloop/state/gates/T1/gate.ps1
if ($LASTEXITCODE -ne 0) {
  exit $LASTEXITCODE
}
& pwsh -NoProfile -File .vloop/state/gates/T2/gate.ps1
if ($LASTEXITCODE -ne 0) {
  exit $LASTEXITCODE
}

# The driver sets VLOOP_PLAN_SHA256 for this gate; the vloop it runs in its own
# scratch repository must not inherit it, or it refuses that repository's plan.
Remove-Item Env:VLOOP_PLAN_SHA256 -ErrorAction SilentlyContinue

$temp = Join-Path ([System.IO.Path]::GetTempPath()) ("vloop-b43-" + [guid]::NewGuid().ToString('N'))
$binary = Join-Path $temp 'vloop.exe'
$repo = Join-Path $temp 'repo'

function Need($condition, $message) {
  if (-not $condition) {
    throw $message
  }
}

function Invoke-Vloop([string[]] $arguments) {
  $start = [System.Diagnostics.ProcessStartInfo]::new()
  $start.FileName = $binary
  $start.WorkingDirectory = $repo
  $start.Arguments = $arguments -join ' '
  $start.UseShellExecute = $false
  $start.RedirectStandardOutput = $true
  $start.RedirectStandardError = $true
  $start.StandardOutputEncoding = [System.Text.Encoding]::UTF8
  $start.StandardErrorEncoding = [System.Text.Encoding]::UTF8
  $process = [System.Diagnostics.Process]::new()
  $process.StartInfo = $start
  [void] $process.Start()
  $stdout = $process.StandardOutput.ReadToEnd()
  $stderr = $process.StandardError.ReadToEnd()
  $process.WaitForExit()
  [pscustomobject]@{
    Code = $process.ExitCode
    Out = $stdout
    Err = $stderr
  }
}

try {
  [void] (New-Item -ItemType Directory -Force -Path $temp)
  [void] (New-Item -ItemType Directory -Force -Path $repo)
  & go build -o $binary ./cmd/vloop
  Need ($LASTEXITCODE -eq 0) "built CLI failed"

  & git -C $repo init --quiet
  & git -C $repo config user.name Gate
  & git -C $repo config user.email gate@example.invalid
  Set-Content -Path (Join-Path $repo 'input.txt') -Value 'fixture' -Encoding utf8
  & git -C $repo add input.txt
  & git -C $repo commit --quiet -m 'café'
  Need ($LASTEXITCODE -eq 0) "fixture commit failed"

  [void] (New-Item -ItemType Directory -Force -Path (Join-Path $repo '.vloop/state'))
  Set-Content -Path (Join-Path $repo '.vloop/config.toml') -Value @(
    'shell = "pwsh"',
    '',
    '[[check]]',
    'name = "gate"',
    'paths = ["**"]',
    'run = "go test ./..."'
  ) -Encoding utf8

  $native = "if (-not (git log -1 --format=%s | Select-String -SimpleMatch 'café' -Quiet)) { exit 1 }"
  $plan = @{
    schema = 'state/v2'
    run_id = 'B20261126-1034-b43-i162-pwsh-gates-utf8'
    brief = 'docs/briefs/B20261126-1034-b43-i162-pwsh-gates-utf8.loop-brief.md'
    base = 'acfa8dc01f72688c17ea28dc6997768236398cd3'
    branch = 'work'
    status = 'planning'
    iteration = 0
    created = '2026-11-26T13:39:18Z'
    updated = '2026-11-26T13:39:18Z'
    shell = 'pwsh'
    checks = @()
    gate_scratch = @()
    gate_review = @{ rounds = 0; verdict = '' }
    tasks = @(
      @{
        id = 'T1'
        title = 'Native read'
        goal = 'Read native UTF-8 output correctly.'
        kind = 'test'
        fixtures = ''
        references = @()
        depends_on = @()
        acceptance = @('The native read passes.')
        verify = $native
        status = 'pending'
        attempts = 0
        notes = ''
      },
      @{
        id = 'T2'
        title = 'UTF-8 output'
        goal = 'Write UTF-8 output correctly.'
        kind = 'test'
        fixtures = ''
        references = @()
        depends_on = @()
        acceptance = @('The UTF-8 output is visible.')
        verify = "Write-Output 'café'"
        status = 'pending'
        attempts = 0
        notes = ''
      },
      @{
        id = 'T3'
        title = 'Exit code'
        goal = 'Preserve the gate exit code.'
        kind = 'test'
        fixtures = ''
        references = @()
        depends_on = @()
        acceptance = @('Exit 3 remains exit 3.')
        verify = 'exit 3'
        status = 'pending'
        attempts = 0
        notes = ''
      }
    )
  }
  $plan | ConvertTo-Json -Depth 8 | Set-Content -Path (Join-Path $repo '.vloop/state/state.json') -Encoding utf8

  $read = Invoke-Vloop @('task', 'gate', 'T1')
  Need ($read.Code -eq 0) "T1 exited $($read.Code): $($read.Out)$($read.Err)"
  Need ($read.Err.Length -eq 0) "T1 wrote stderr: $($read.Err)"

  $write = Invoke-Vloop @('task', 'gate', 'T2')
  Need ($write.Code -eq 0) "T2 exited $($write.Code): $($write.Out)$($write.Err)"
  Need ($write.Out -match '(?m)^café\r?$') "T2 did not emit café: $($write.Out)"
  Need ($write.Err.Length -eq 0) "T2 wrote stderr: $($write.Err)"

  $failure = Invoke-Vloop @('task', 'gate', 'T3')
  Need ($failure.Code -eq 1) "T3 command exited $($failure.Code)"
  Need ($failure.Out -match 'gate T3: fail \(exit 3,') "T3 did not report exit 3: $($failure.Out)"
  Need ($failure.Err.Length -eq 0) "T3 wrote stderr: $($failure.Err)"

  $shown = Invoke-Vloop @('task', 'show', 'T1')
  Need ($shown.Code -eq 0) "task show exited $($shown.Code)"
  Need ($shown.Out.Contains("verify  $native")) "task show changed the recorded command: $($shown.Out)"
  Need (-not $shown.Out.Contains('[Console]::OutputEncoding')) "task show exposed the setup"
  Need (-not $shown.Out.Contains('[Console]::InputEncoding')) "task show exposed the setup"
  Need (-not $shown.Out.Contains('$OutputEncoding')) "task show exposed the setup"
} finally {
  Remove-Item -Recurse -Force -Path $temp -ErrorAction SilentlyContinue
}
