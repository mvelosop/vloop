. (Join-Path $PSScriptRoot "..\T1\harness.ps1")

$h = New-GateHarness -Version "0.15.0" -Module "v0.15.0" -Agents @("claude", "copilot")
try {
    $env:FAKE_CLAUDE_LIST = '[{"id":"vloop@vloop","version":"0.14.0","scope":"user","enabled":true}]'
    $env:FAKE_COPILOT_LIST = '[{"name":"vloop","version":"0.14.0","enabled":true}]'
    $env:Path = $h.Fake + [IO.Path]::PathSeparator + $script:OriginalPath
    $env:GH_TOKEN = "gate-token"

    $help = Invoke-GateCommand $h.Exe @("upgrade", "--help")
    Assert-Equal $help.Code 0 "upgrade help exits 0"
    Assert-Gate ($help.Out.Contains("To update the vloop binary and its plugins, run vloop self-update first.")) "upgrade help points to self-update"

    # Doctor builds the claude plugin row only when a session runs on claude, and
    # the copilot one only when one runs on copilot: name one of each through the
    # environment, which beats any .vloop/config.local.toml (C-1).
    $env:VLOOP_AGENT_WORK = "claude"; $env:VLOOP_MODEL_WORK = "sonnet"
    $env:VLOOP_AGENT_REVIEW = "copilot"; $env:VLOOP_MODEL_REVIEW = "gpt-6.1-terra"
    $doctor = Invoke-GateCommand $h.Exe @("--json", "doctor")
    Remove-Item Env:VLOOP_AGENT_WORK, Env:VLOOP_MODEL_WORK, Env:VLOOP_AGENT_REVIEW, Env:VLOOP_MODEL_REVIEW -ErrorAction SilentlyContinue
    $checks = ($doctor.Out | ConvertFrom-Json).checks
    $claude = $checks | Where-Object { $_.name -eq "plugin" } | Select-Object -Last 1
    $copilot = $checks | Where-Object { $_.name -eq "copilot plugin" } | Select-Object -Last 1
    Assert-Gate ($claude.message.EndsWith(" — vloop run is unaffected; vloop self-update updates the binary and the plugin")) "doctor claude self-update pointer"
    Assert-Gate ($copilot.message.EndsWith(" — vloop run is unaffected; vloop self-update updates the binary and the plugin")) "doctor copilot self-update pointer"

    $conceptParagraphs = (Get-Content (Join-Path $script:RepoRoot "docs\guide\concepts.md") -Raw) -split "`n`n"
    $selfUpdate = $conceptParagraphs | Where-Object { $_ -match "vloop self-update" } | Select-Object -First 1
    Assert-Gate ($null -ne $selfUpdate) "concepts explains self-update"
    Assert-Gate (($selfUpdate -match "checkout") -and ($selfUpdate -match "go") -and ($selfUpdate -match "PATH")) "concepts names self-update refusals in one paragraph"

    $commands = Get-Content (Join-Path $script:RepoRoot "docs\guide\commands.md") -Raw
    Assert-Gate ($commands.Contains("vloop self-update")) "generated commands includes self-update"

    $domain = (Get-Content (Join-Path $script:RepoRoot "docs\domain\domain-model.md") -Raw) -replace '\s+', ' '
    $platform = (Get-Content (Join-Path $script:RepoRoot "docs\domain\platform\platform-context.md") -Raw) -replace '\s+', ' '
    Assert-Gate ($domain.Contains("~/")) "domain C-2 documents home shorthand"
    Assert-Gate ($domain.Contains("outside") -and $domain.Contains("repository")) "domain C-2 documents outside-repository paths"
    Assert-Gate ($platform.Contains("self-update") -and $platform.Contains("~/")) "platform context documents self-update path exception"

    $release = Join-Path $h.Temp "release.exe"
    & $script:RealGo build -o $release .\tools\release
    Assert-Equal $LASTEXITCODE 0 "release tool builds"
    & $script:RealGo build -o (Join-Path $h.Temp "faketool.exe") (Join-Path $PSScriptRoot "faketool.go")
    Assert-Equal $LASTEXITCODE 0 "release fake builds"
    $gitName = if ([OperatingSystem]::IsWindows()) { "git.exe" } else { "git" }
    $ghName = if ([OperatingSystem]::IsWindows()) { "gh.exe" } else { "gh" }
    Copy-Item (Join-Path $h.Temp "faketool.exe") (Join-Path $h.Fake $gitName)
    Copy-Item (Join-Path $h.Temp "faketool.exe") (Join-Path $h.Fake $ghName)
    $savedGobin = $env:FAKE_GOBIN
    $env:FAKE_GOBIN = ""
    $env:FAKE_GO_LATEST = "v0.15.0"
    $env:FAKE_GO_INSTALL_EXE = $h.Exe
    $published = Invoke-GateCommand $release @("publish", "v0.15.0")
    $env:FAKE_GOBIN = $savedGobin
    Assert-Equal $published.Code 0 "release publish proof exits 0"
    $last = ($published.Out -split "`n")[-1]
    Assert-Equal $last "then: vloop self-update, and vloop upgrade in each repository" "release handoff line"
} finally {
    Remove-Item $h.Temp -Recurse -Force -ErrorAction SilentlyContinue
    $env:Path = $script:OriginalPath
    Remove-Item Env:GH_TOKEN -ErrorAction SilentlyContinue
}
