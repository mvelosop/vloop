$ErrorActionPreference = 'Stop'

foreach ($task in @('T1', 'T2', 'T3', 'T4')) {
  & pwsh -NoProfile -File ".vloop/state/gates/$task/gate.ps1"
  if ($LASTEXITCODE -ne 0) { throw "$task gate failed" }
}

$root = (Get-Location).Path
$temp = Join-Path ([System.IO.Path]::GetTempPath()) ("vloop-gate-t5-" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $temp | Out-Null
try {
  $binary = Join-Path $temp 'vloop.exe'
  & go build -o $binary ./cmd/vloop
  if ($LASTEXITCODE -ne 0) { throw 'failed to build vloop' }

  function Assert-Fence {
    param([string]$Tool, [string]$Command, [string]$Expected)
    $output = (& $binary fence check --phase work --root $root --tool $Tool $Command 2>&1 | Out-String).Trim()
    $code = $LASTEXITCODE
    $wantedCode = if ($Expected -eq 'allow') { 0 } else { 1 }
    if ($code -ne $wantedCode -or $output -ne $Expected) {
      throw "[$Tool] $Command`nexpected ($wantedCode) $Expected`nactual   ($code) $output"
    }
  }

  $cases = @(
    [pscustomobject]@{ Tool = 'powershell'; Command = "pwsh -c 'git push'"; Expected = 'deny: git push is denied by the work fence (Bash(git push:*))' },
    [pscustomobject]@{ Tool = 'bash'; Command = "pwsh -NoProfile -Command `"git push origin main`""; Expected = 'deny: git push origin main is denied by the work fence (Bash(git push:*))' },
    [pscustomobject]@{ Tool = 'powershell'; Command = "pwsh -NoProfile -Command 'go test ./x'"; Expected = 'allow' },
    [pscustomobject]@{ Tool = 'powershell'; Command = 'pwsh -File x.ps1'; Expected = 'deny: pwsh: runs a script, stdin or an encoded command, not a literal -Command string' },
    [pscustomobject]@{ Tool = 'powershell'; Command = 'powershell -EncodedCommand ZwBpAHQA'; Expected = 'deny: powershell: runs a script, stdin or an encoded command, not a literal -Command string' },
    [pscustomobject]@{ Tool = 'powershell'; Command = 'pwsh -c $x'; Expected = 'deny: the -Command string expands and cannot be judged' },
    [pscustomobject]@{ Tool = 'powershell'; Command = 'cmd /c git push'; Expected = 'deny: git push is denied by the work fence (Bash(git push:*))' },
    [pscustomobject]@{ Tool = 'powershell'; Command = 'cmd /c echo %PATH%'; Expected = 'deny: cmd: % and ^ cannot be judged' },
    [pscustomobject]@{ Tool = 'powershell'; Command = 'cmd'; Expected = 'deny: cmd: runs stdin or a batch file, not a literal /c string' },
    [pscustomobject]@{ Tool = 'powershell'; Command = "iex 'git push'"; Expected = 'deny: iex is not allowed' },
    [pscustomobject]@{ Tool = 'powershell'; Command = 'Get-Content a | Invoke-Expression'; Expected = 'deny: Invoke-Expression is not allowed' },
    [pscustomobject]@{ Tool = 'powershell'; Command = 'Start-Process git push'; Expected = 'deny: Start-Process is not allowed' },
    [pscustomobject]@{ Tool = 'powershell'; Command = 'Get-ChildItem | ForEach-Object { $_.Name }'; Expected = 'deny: script blocks cannot be judged — use Where-Object <Property> -like <value>, ForEach-Object -MemberName <Property> or Select-Object -ExpandProperty <Property>' },
    [pscustomobject]@{ Tool = 'powershell'; Command = '& {git push}'; Expected = 'deny: script blocks cannot be judged — use Where-Object <Property> -like <value>, ForEach-Object -MemberName <Property> or Select-Object -ExpandProperty <Property>' },
    [pscustomobject]@{ Tool = 'powershell'; Command = 'Get-ChildItem | ForEach-Object -MemberName Name'; Expected = 'allow' },
    [pscustomobject]@{ Tool = 'powershell'; Command = '(Get-Content README.md | Measure-Object -Line).Lines'; Expected = 'allow' },
    [pscustomobject]@{ Tool = 'powershell'; Command = '(git push).Length'; Expected = 'deny: git push is denied by the work fence (Bash(git push:*))' },
    [pscustomobject]@{ Tool = 'powershell'; Command = '(Get-Content a).Trim()'; Expected = 'deny: parentheses cannot be judged here — (<command>).<Property> is read; or pipe into Measure-Object or Select-Object -ExpandProperty' },
    [pscustomobject]@{ Tool = 'powershell'; Command = '$env:GOOS=''linux''; go build ./...'; Expected = 'allow' },
    [pscustomobject]@{ Tool = 'powershell'; Command = '$env:GOOS = "linux"; go vet ./...'; Expected = 'allow' },
    [pscustomobject]@{ Tool = 'powershell'; Command = '$env:GOOS=$x; go build ./...'; Expected = 'deny: command word $env:GOOS=$x starts with $ and cannot be judged' },
    [pscustomobject]@{ Tool = 'powershell'; Command = 'Write-Output "a`tb"'; Expected = 'deny: a backtick cannot be judged — write the command on one line, without escapes' },
    [pscustomobject]@{ Tool = 'powershell'; Command = 'Write-Output "$(git push)"'; Expected = 'deny: command substitution cannot be judged — run the inner command on its own' },
    [pscustomobject]@{ Tool = 'bash'; Command = '(cat a).Count'; Expected = 'deny: subshells and process substitution cannot be judged' },
    [pscustomobject]@{ Tool = 'bash'; Command = 'eval "git push"'; Expected = 'deny: eval is not allowed' },
    [pscustomobject]@{ Tool = 'bash'; Command = "bash -c 'git push'"; Expected = 'deny: git push is denied by the work fence (Bash(git push:*))' }
  )
  foreach ($case in $cases) {
    Assert-Fence $case.Tool $case.Command $case.Expected
  }
}
finally {
  Remove-Item -LiteralPath $temp -Recurse -Force
}
