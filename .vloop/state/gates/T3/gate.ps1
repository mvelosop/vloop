. (Join-Path $PSScriptRoot "..\T1\harness.ps1")

$h = New-GateHarness -Latest "v0.14.0" -Agents @("claude", "copilot")
try {
    $env:FAKE_CLAUDE_LIST = '[{"id":"vloop@vloop","version":"0.14.0","scope":"","enabled":true},{"id":"vloop@vloop","version":"0.14.0","scope":"project","projectPath":"' + $script:RepoRoot.Replace('\', '/') + '","enabled":true},{"id":"vloop@vloop","version":"0.14.0","scope":"local","projectPath":"/somewhere-else","enabled":true}]'
    $env:FAKE_COPILOT_LIST = '[{"name":"vloop","version":"0.14.0","enabled":true}]'
    $r = Invoke-GateCommand $h.Exe @("self-update")
    $want = @(
        "vloop v0.14.0 is up to date (latest release v0.14.0)"
        "plugin claude (user): updated"
        "plugin claude (project): updated"
        "plugin copilot: updated"
        "next: vloop upgrade in each repository that uses vloop, and vloop self-update there for a plugin installed for that repository"
    ) -join "`n"
    Assert-Equal $r.Code 0 "plugin updates exit 0"
    Assert-Equal $r.Out $want "plugin success output"
    $log = Get-GateLog $h.AgentLog
    Assert-Gate ($log.Contains("claude plugin marketplace update vloop")) "claude marketplace refreshed first"
    Assert-Gate ($log.Contains("claude plugin update vloop@vloop --scope user")) "claude user scope updated"
    Assert-Gate ($log.Contains("claude plugin update vloop@vloop --scope project")) "claude project scope updated"
    Assert-Gate (-not $log.Contains(" --scope local")) "foreign project installation is ignored"
    Assert-Gate ($log.Contains("copilot plugin update vloop@vloop")) "copilot updated"

    $env:FAKE_CLAUDE_UPDATE_FAIL = "network down"
    $env:FAKE_COPILOT_UPDATE_FAIL = "copilot denied"
    $r = Invoke-GateCommand $h.Exe @("self-update")
    Assert-Equal $r.Code 0 "plugin failures remain non-fatal"
    Assert-Gate ($r.Out.Contains("! plugin claude (user): claude plugin update vloop@vloop --scope user failed: network down — run it by hand")) "claude failure line"
    Assert-Gate ($r.Out.Contains("! plugin copilot: copilot plugin update vloop@vloop failed: copilot denied — run it by hand")) "copilot failure line"
    $env:FAKE_CLAUDE_UPDATE_FAIL = ""
    $env:FAKE_COPILOT_UPDATE_FAIL = ""

    $env:FAKE_CLAUDE_MARKETPLACE_FAIL = "market unavailable"
    $r = Invoke-GateCommand $h.Exe @("self-update")
    Assert-Equal $r.Code 0 "marketplace failure remains non-fatal"
    Assert-Gate ($r.Out.Contains("! plugin claude: claude plugin marketplace update vloop failed: market unavailable — run it by hand")) "marketplace failure line"
    Assert-Gate (-not $r.Out.Contains("plugin claude (user): updated")) "marketplace failure skips update"
    $env:FAKE_CLAUDE_MARKETPLACE_FAIL = ""

    $env:FAKE_CLAUDE_LIST = "[]"
    $env:FAKE_COPILOT_LIST = "[]"
    $r = Invoke-GateCommand $h.Exe @("self-update")
    Assert-Gate ($r.Out.Contains("plugin claude: not installed")) "claude not installed line"
    Assert-Gate ($r.Out.Contains("plugin copilot: not installed")) "copilot not installed line"

    $beforeCheck = if (Test-Path $h.AgentLog) { (Get-Item $h.AgentLog).Length } else { 0 }
    $r = Invoke-GateCommand $h.Exe @("self-update", "--check")
    $afterCheck = if (Test-Path $h.AgentLog) { (Get-Item $h.AgentLog).Length } else { 0 }
    Assert-Equal $afterCheck $beforeCheck "--check does not inspect plugins"

    $env:FAKE_CLAUDE_LIST = '[{"id":"vloop@vloop","version":"0.14.0","scope":"user","enabled":true}]'
    $env:FAKE_COPILOT_LIST = "not-json"
    $r = Invoke-GateCommand $h.Exe @("self-update")
    Assert-Gate ($r.Out.Contains("! plugin copilot: cannot read copilot plugin list --json:")) "invalid list is reported"
    Assert-Gate ($r.Out.Contains(" — run copilot plugin update vloop@vloop by hand")) "invalid list remedy"
    $env:FAKE_COPILOT_LIST = "[]"

    $copilotName = if ([OperatingSystem]::IsWindows()) { "copilot.exe" } else { "copilot" }
    Remove-Item (Join-Path $h.Fake $copilotName) -Force
    $r = Invoke-GateCommand $h.Exe @("--json", "self-update")
    Assert-Equal $r.Code 0 "plugin JSON exits 0"
    $doc = $r.Out | ConvertFrom-Json
    Assert-Equal $doc.plugins.Count 1 "plugin JSON entries"
    Assert-Equal $doc.plugins[0].agent "claude" "plugin JSON claude"
    Assert-Equal $doc.plugins[0].scope "user" "plugin JSON scope"
    Assert-Equal $doc.plugins[0].result "updated" "plugin JSON result"
    Assert-Equal $doc.plugins[0].message "" "plugin JSON message"

    $r = Invoke-GateCommand $h.Exe @("self-update", "--version", "v0.14.0")
    Assert-Gate ($r.Out.Contains("! with --version, the plugins still update to their latest release, which may differ from vloop v0.14.0")) "human --version plugin warning"

    $env:FAKE_CLAUDE_UPDATE_FAIL = "network down"
    $r = Invoke-GateCommand $h.Exe @("--json", "self-update", "--version", "v0.13.0")
    Assert-Equal $r.Code 0 "JSON with version exits 0"
    Assert-Gate (-not $r.Out.Contains("with --version")) "JSON suppresses version warning"
    Assert-Equal (($r.Out | ConvertFrom-Json).plugins[0].message) "network down" "JSON failure message"

} finally {
    Remove-Item $h.Temp -Recurse -Force -ErrorAction SilentlyContinue
    $env:Path = $script:OriginalPath
}
