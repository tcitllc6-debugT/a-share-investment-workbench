param(
  [Parameter(Mandatory = $true)][string]$Version,
  [Parameter(Mandatory = $true)][string]$ManifestKeyPath,
  [string]$BaseUrl = "",
  [string]$Notes = "自动构建更新"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
if ([string]::IsNullOrWhiteSpace($BaseUrl)) {
  $BaseUrl = "https://github.com/tcitllc6-debugT/a-share-investment-workbench/releases/download/v$Version"
}

$updater = Join-Path $root "build\bin\a-share-updater.exe"
$application = Get-ChildItem (Join-Path $root "build\bin") -Filter "*.exe" -File |
  Where-Object { $_.Name -ne "a-share-updater.exe" -and $_.Name -notlike "*-installer.exe" } |
  Sort-Object LastWriteTime -Descending |
  Select-Object -First 1
foreach ($file in @($application.FullName, $updater, $ManifestKeyPath)) {
  if (-not (Test-Path -LiteralPath $file)) { throw "缺少发布所需文件：$file" }
}

$releaseDir = Join-Path $root "build\release\$Version"
$packageDir = Join-Path $releaseDir "package"
New-Item -ItemType Directory -Force -Path $packageDir | Out-Null
Copy-Item -LiteralPath $application.FullName -Destination (Join-Path $packageDir $application.Name) -Force
Copy-Item -LiteralPath $updater -Destination (Join-Path $packageDir "a-share-updater.exe") -Force

$assetName = "a-share-investment-workbench-windows-amd64.zip"
$assetPath = Join-Path $releaseDir $assetName
Remove-Item -LiteralPath $assetPath -Force -ErrorAction SilentlyContinue
Compress-Archive -Path (Join-Path $packageDir "*") -DestinationPath $assetPath -CompressionLevel Optimal
$hash = (Get-FileHash -LiteralPath $assetPath -Algorithm SHA256).Hash.ToLowerInvariant()

$manifest = [ordered]@{
  version = $Version
  publishedAt = (Get-Date).ToUniversalTime().ToString("o")
  notes = $Notes
  assets = @([ordered]@{
    name = $assetName
    os = "windows"
    arch = "amd64"
    url = "$BaseUrl/$assetName"
    sha256 = $hash
    size = (Get-Item -LiteralPath $assetPath).Length
  })
  signature = ""
}
$manifestPath = Join-Path $releaseDir "manifest.json"
[IO.File]::WriteAllText($manifestPath, ($manifest | ConvertTo-Json -Depth 8), [Text.UTF8Encoding]::new($false))
go run .\backend\cmd\update-manifest sign -key $ManifestKeyPath -manifest $manifestPath

Write-Host "更新包：$assetPath"
Write-Host "签名清单：$manifestPath"
