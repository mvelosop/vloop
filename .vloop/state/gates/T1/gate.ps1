. (Join-Path $PSScriptRoot "harness.ps1")

$h = New-GateHarness
$h2 = $null
try {
    $r = Invoke-GateCommand $h.Exe @("self-update", "--check")
    Assert-Equal $r.Code 0 "newer --check exits 0"
    Assert-Equal $r.Out "current  v0.14.0`nlatest   v0.15.0 — vloop self-update installs it" "newer --check output"
    Assert-Equal $r.Err "" "newer --check stderr"

    $r = Invoke-GateCommand $h.Exe @("self-update", "--check", "--version", "v0.14.0")
    Assert-Equal $r.Out "current  v0.14.0`ntarget   v0.14.0 — already running" "--check --version equal output"

    $r = Invoke-GateCommand $h.Exe @("self-update", "--version", "0.15")
    Assert-Equal $r.Code 2 "invalid --version exits 2"
    Assert-Equal $r.Out "" "invalid --version stdout"
    Assert-Equal $r.Err "vloop: --version 0.15 is not a version like v0.15.0" "invalid --version message"

    $env:FAKE_GO_LIST_FAIL = "lookup broke"
    $r = Invoke-GateCommand $h.Exe @("self-update", "--check")
    Assert-Equal $r.Code 1 "failed latest lookup exits 1"
    Assert-Equal $r.Err "vloop: cannot find the latest release: first line" "failed latest lookup message"
    $env:FAKE_GO_LIST_FAIL = ""

    $env:VLOOP_GATE_MODULE = ""
    $r = Invoke-GateCommand $h.Exe @("self-update", "--check")
    Assert-Equal $r.Code 1 "checkout build exits 1"
    Assert-Equal $r.Err "vloop: this vloop was built from a checkout, not installed from a release — build it again from the checkout, or go install github.com/mvelosop/vloop/cmd/vloop@latest" "checkout build message"
    $env:VLOOP_GATE_MODULE = "v0.14.0"

    $savedPath = $env:Path
    $empty = Join-Path $h.Temp "empty"
    New-Item -ItemType Directory -Force $empty | Out-Null
    $env:Path = $empty
    $r = Invoke-GateCommand $h.Exe @("self-update")
    $env:Path = $savedPath
    Assert-Equal $r.Code 1 "missing go exits 1"
    Assert-Equal $r.Err "vloop: self-update needs go on PATH — install the release by hand: go install github.com/mvelosop/vloop/cmd/vloop@latest" "missing go message"

    $h2 = New-GateHarness -ExeOutsideGobin
    $r = Invoke-GateCommand $h2.Exe @("self-update")
    Assert-Equal $r.Code 1 "wrong install location exits 1"
    $want = "vloop: this vloop runs from $(Format-GatePath $h2.Exe), but go install writes to $(Format-GatePath $h2.Gobin) — update it the way you installed it, or put $(Format-GatePath $h2.Gobin) first on PATH"
    Assert-Equal $r.Err $want "wrong install location message"

    $h3 = New-GateHarness
    $env:FAKE_GO_ENV_FAIL = "go env broke"
    $r = Invoke-GateCommand $h3.Exe @("self-update")
    Assert-Equal $r.Code 1 "go env failure exits 1"
    Assert-Equal $r.Err "vloop: cannot find where go install writes: go env broke" "go env failure message"
    $env:FAKE_GO_ENV_FAIL = ""
    $env:FAKE_GOBIN = ""
    $env:FAKE_GOPATH = ""
    $r = Invoke-GateCommand $h3.Exe @("self-update")
    Assert-Equal $r.Err "vloop: cannot find where go install writes: GOBIN and GOPATH are empty" "empty go install directory message"

    $env:VLOOP_GATE_MODULE = "v0.15.0-rc.4"
    $env:FAKE_GO_LATEST = "v0.14.0"
    $env:FAKE_GO_ENV_FAIL = "must not be called"
    $r = Invoke-GateCommand $h2.Exe @("self-update")
    Assert-Equal $r.Out "vloop v0.15.0-rc.4 is up to date (latest release v0.14.0)" "prerelease newer than latest output"

    $env:VLOOP_GATE_MODULE = "v0.14.0"
    $env:FAKE_GO_LATEST = "v0.15.0"
    $env:FAKE_GO_ENV_FAIL = ""
    $r = Invoke-GateCommand $h.Exe @("--json", "self-update", "--check")
    Assert-Equal $r.Code 0 "JSON --check exits 0"
    $doc = $r.Out | ConvertFrom-Json
    Assert-Equal $doc.schema "self-update/v1" "JSON schema name"
    Assert-Equal $doc.current "v0.14.0" "JSON current"
    Assert-Equal $doc.target "v0.15.0" "JSON target"
    Assert-Gate ($doc.newer -eq $true) "JSON newer"
    Assert-Gate ($doc.installed -eq $false) "JSON installed"
    Assert-Equal $doc.dir "" "JSON check dir"
    Assert-Equal $doc.plugins.Count 0 "JSON check plugins"
    $jsonFile = Join-Path $h.Temp "result.json"
    [IO.File]::WriteAllText($jsonFile, $r.Out)
    $schema = Invoke-GateCommand $h.Exe @("schema", "validate", "self-update/v1", $jsonFile)
    Assert-Equal $schema.Code 0 "self-update JSON validates against its embedded schema"
} finally {
    Remove-Item $h.Temp -Recurse -Force -ErrorAction SilentlyContinue
    if ($h2) { Remove-Item $h2.Temp -Recurse -Force -ErrorAction SilentlyContinue }
    if ($h3) { Remove-Item $h3.Temp -Recurse -Force -ErrorAction SilentlyContinue }
    $env:Path = $script:OriginalPath
}
