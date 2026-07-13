param(
  [Parameter(Mandatory = $true)][string]$Version,
  [string]$BaseUrl = "",
  [string]$Notes = "稳定性与选股功能更新"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
if ([string]::IsNullOrWhiteSpace($BaseUrl)) {
  $BaseUrl = "https://github.com/tcitllc6-debugT/a-share-investment-workbench/releases/download/v$Version"
}
$application = Get-ChildItem (Join-Path $root "build\bin") -Filter "*.exe" -File |
  Where-Object { $_.Name -ne "a-share-updater.exe" -and $_.Name -notlike "*-installer.exe" } |
  Sort-Object LastWriteTime -Descending |
  Select-Object -First 1
if (-not $application) { throw "未找到主程序，请先运行 scripts\build-windows.ps1" }

$releaseDir = Join-Path $root "build\release\$Version"
$stagingDir = Join-Path $releaseDir "package"
New-Item -ItemType Directory -Force -Path $stagingDir | Out-Null
Copy-Item $application.FullName (Join-Path $stagingDir $application.Name) -Force

$assetName = "a-share-investment-workbench-windows-amd64.zip"
$assetPath = Join-Path $releaseDir $assetName
Remove-Item $assetPath -Force -ErrorAction SilentlyContinue
Compress-Archive -Path (Join-Path $stagingDir "*") -DestinationPath $assetPath -CompressionLevel Optimal
$hash = (Get-FileHash -Path $assetPath -Algorithm SHA256).Hash.ToLowerInvariant()
$size = (Get-Item $assetPath).Length

$manifest = [ordered]@{
  version = $Version
  publishedAt = (Get-Date).ToUniversalTime().ToString("o")
  notes = $Notes
  assets = @(
    [ordered]@{
      name = $assetName
      os = "windows"
      arch = "amd64"
      url = "$BaseUrl/$assetName"
      sha256 = $hash
      size = $size
    }
  )
  signature = ""
}
$manifestPath = Join-Path $releaseDir "manifest.json"
$manifestJson = $manifest | ConvertTo-Json -Depth 8
[System.IO.File]::WriteAllText($manifestPath, $manifestJson, [System.Text.UTF8Encoding]::new($false))
go run .\backend\cmd\update-manifest sign -key .\.update-signing\manifest-private.key -manifest $manifestPath
Write-Host "发布目录：$releaseDir"
Write-Host "签名清单：$manifestPath"
