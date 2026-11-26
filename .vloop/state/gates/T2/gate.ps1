. (Join-Path $PSScriptRoot "..\T1\harness.ps1")

$h = New-GateHarness
$hOld = $null
try {
    New-Item -ItemType File -Force (Join-Path $h.Gobin ($h.ExeName + ".old")) | Out-Null
    $r = Invoke-GateCommand $h.Exe @("self-update")
    Assert-Equal $r.Code 0 "successful install exits 0"
    $want = "installing vloop v0.15.0 (from v0.14.0)…`ninstalled vloop v0.15.0 at $(Format-GatePath $h.Gobin)`nnext: vloop upgrade in each repository that uses vloop, and vloop self-update there for a plugin installed for that repository"
    Assert-Equal $r.Out $want "successful install output"
    Assert-Equal $r.Err "" "successful install stderr"
    Assert-Equal (Get-GateLog $h.GoLog) "install github.com/mvelosop/vloop/cmd/vloop@v0.15.0" "install command"

    if ([OperatingSystem]::IsWindows()) {
        $hOld = New-GateHarness
        $old = Join-Path $hOld.Gobin ($hOld.ExeName + ".old")
        New-Item -ItemType Directory -Force $old | Out-Null
        [IO.File]::WriteAllText((Join-Path $old "locked"), "keep")
        $r = Invoke-GateCommand $hOld.Exe @("self-update")
        Assert-Equal $r.Code 1 "unremovable old executable exits 1"
        Assert-Gate ($r.Err.StartsWith("vloop: cannot remove $(Format-GatePath $old) (")) "old removal failure prefix"
        Assert-Gate ($r.Err.EndsWith(") — delete it and run vloop self-update again")) "old removal failure remedy"
        Assert-Gate (-not (Test-Path $hOld.GoLog)) "old removal failure installs nothing"
    }

    $h2 = New-GateHarness
    $env:FAKE_GO_INSTALL_FAIL = "no matching versions"
    $r = Invoke-GateCommand $h2.Exe @("self-update")
    Assert-Equal $r.Code 1 "failed install exits 1"
    Assert-Equal $r.Err "vloop: go install failed: no matching versions" "failed install message"
    Assert-Gate (Test-Path $h2.Exe) "failed install restores the executable"
    $probe = Invoke-GateCommand $h2.Exe @("version", "--json")
    Assert-Equal $probe.Code 0 "restored executable still runs"
    $version = $probe.Out | ConvertFrom-Json
    Assert-Equal $version.module "v0.14.0" "restored executable keeps old version"

    $h3 = New-GateHarness
    $r = Invoke-GateCommand $h3.Exe @("--json", "self-update", "--version", "v0.13.0")
    Assert-Equal $r.Code 0 "explicit older install exits 0"
    $doc = $r.Out | ConvertFrom-Json
    Assert-Equal $doc.schema "self-update/v1" "install JSON schema"
    Assert-Equal $doc.target "v0.13.0" "install JSON target"
    Assert-Gate ($doc.newer -eq $false) "explicit older JSON newer"
    Assert-Gate ($doc.installed -eq $true) "install JSON installed"
    Assert-Equal $doc.dir (Format-GatePath $h3.Gobin) "install JSON dir"
    Assert-Equal (Get-GateLog $h3.GoLog) "install github.com/mvelosop/vloop/cmd/vloop@v0.13.0" "explicit target install command"
} finally {
    Remove-Item $h.Temp -Recurse -Force -ErrorAction SilentlyContinue
    if ($hOld) { Remove-Item $hOld.Temp -Recurse -Force -ErrorAction SilentlyContinue }
    if ($h2) { Remove-Item $h2.Temp -Recurse -Force -ErrorAction SilentlyContinue }
    if ($h3) { Remove-Item $h3.Temp -Recurse -Force -ErrorAction SilentlyContinue }
    $env:Path = $script:OriginalPath
}
