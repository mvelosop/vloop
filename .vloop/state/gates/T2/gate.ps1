$ErrorActionPreference = 'Stop'
$path = 'docs/guide/configuration.md'
$line = '[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new()'
if (-not (Select-String -Path $path -SimpleMatch $line -Quiet)) {
  Write-Error "missing PowerShell profile line: $line"
  exit 1
}
