param([string]$CoreRepo = "..\KINETOVELA")
$ErrorActionPreference = "Stop"
$manifest = Get-Content (Join-Path $PSScriptRoot "..\contracts\sync-manifest.json") -Raw | ConvertFrom-Json
$failed = $false
foreach ($entry in $manifest.files) {
  $src = Join-Path $CoreRepo ("contracts/" + $entry.path)
  $dst = Join-Path $PSScriptRoot ("../contracts/" + $entry.path)
  if (-not (Test-Path $src)) { Write-Error "missing source: $src"; $failed = $true; continue }
  $a = (Get-FileHash $src -Algorithm SHA256).Hash.ToLower()
  $b = (Get-FileHash $dst -Algorithm SHA256).Hash.ToLower()
  if ($a -ne $b) { Write-Error "drift: $($entry.path)"; $failed = $true }
}
if ($failed) { exit 1 }
Write-Output "contract sync check passed"
