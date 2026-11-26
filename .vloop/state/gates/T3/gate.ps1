$ErrorActionPreference = 'Stop'
$guide = Get-Content -Raw "docs/guide/concepts.md"
$sentence = '`init` writes `.vloop/README.md`, and `upgrade` rewrites or restores it.'
if (-not $guide.Contains($sentence)) {
    throw "Setting up a repository lacks the README lifecycle sentence"
}
