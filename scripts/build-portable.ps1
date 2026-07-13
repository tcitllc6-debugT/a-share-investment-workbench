param(
  [string]$Version = "1.2.0"
)

$ErrorActionPreference = "Stop"
$root = [System.IO.Path]::GetFullPath((Split-Path -Parent $PSScriptRoot))
$binDir = Join-Path $root "build\bin"
$updater = Join-Path $binDir "a-share-updater.exe"
$webViewInstaller = Join-Path $root "build\windows\installer\tmp\MicrosoftEdgeWebview2Setup.exe"
$application = Get-ChildItem $binDir -Filter "*.exe" -File |
  Where-Object { $_.Name -ne "a-share-updater.exe" -and $_.Name -notlike "*-installer.exe" } |
  Sort-Object LastWriteTime -Descending |
  Select-Object -First 1

foreach ($path in @($application.FullName, $updater)) {
  if ([string]::IsNullOrWhiteSpace($path) -or -not (Test-Path -LiteralPath $path)) {
    throw "Missing release file: $path. Run scripts\build-windows.ps1 first."
  }
}

$portableBase = [System.IO.Path]::GetFullPath((Join-Path $root "build\portable"))
$portableRoot = [System.IO.Path]::GetFullPath((Join-Path $portableBase "A股智投工作台-便携版"))
if (-not $portableRoot.StartsWith($portableBase, [System.StringComparison]::OrdinalIgnoreCase)) {
  throw "Portable output path is outside the workspace."
}

Remove-Item -LiteralPath $portableRoot -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path $portableRoot | Out-Null

Copy-Item -LiteralPath $application.FullName -Destination (Join-Path $portableRoot $application.Name)
Copy-Item -LiteralPath $updater -Destination (Join-Path $portableRoot "a-share-updater.exe")
Copy-Item -LiteralPath (Join-Path $root "LICENSE") -Destination (Join-Path $portableRoot "LICENSE")
Copy-Item -LiteralPath (Join-Path $root "NOTICE") -Destination (Join-Path $portableRoot "NOTICE")
Copy-Item -LiteralPath (Join-Path $root "README.md") -Destination (Join-Path $portableRoot "README.md")
if (Test-Path -LiteralPath $webViewInstaller) {
  Copy-Item -LiteralPath $webViewInstaller -Destination (Join-Path $portableRoot "MicrosoftEdgeWebview2Setup.exe")
}

$launcher = @'
@echo off
for %%F in ("%~dp0*.exe") do if /I not "%%~nxF"=="a-share-updater.exe" (start "" "%%~fF" & exit /b)
'@
[System.IO.File]::WriteAllText((Join-Path $portableRoot "启动 A股智投工作台.cmd"), $launcher, [System.Text.Encoding]::ASCII)

if (Test-Path -LiteralPath $webViewInstaller) {
  $webViewLauncher = @'
@echo off
start "" "%~dp0MicrosoftEdgeWebview2Setup.exe"
'@
  [System.IO.File]::WriteAllText((Join-Path $portableRoot "无法启动时安装 WebView2.cmd"), $webViewLauncher, [System.Text.Encoding]::ASCII)
}

$readme = @'
A股智投工作台便携版

1. 将整个文件夹解压到桌面、D:\Tools 或其他可写目录。
2. 双击“启动 A股智投工作台.cmd”或“A股智投工作台.exe”运行。
3. 不要放入 Program Files：便携版在线更新需要对当前目录具有写入权限。
4. 自选、持仓、交易日志和设置保存在本机数据目录，更新程序不会删除这些数据。
5. 较旧的 Windows 如果无法启动，可运行“无法启动时安装 WebView2.cmd”。
6. 可在“设置 > 检查更新”中从 GitHub Releases 获取签名更新，更新完成后会自动重启。

本项目按 GNU GPL v3.0 发布，详情见 LICENSE 与 NOTICE。
对应源码：https://github.com/tcitllc6-debugT/a-share-investment-workbench
'@
[System.IO.File]::WriteAllText((Join-Path $portableRoot "使用说明.txt"), $readme, [System.Text.UTF8Encoding]::new($true))

$releaseDir = Join-Path $root "build\release\$Version"
New-Item -ItemType Directory -Force -Path $releaseDir | Out-Null
$archive = Join-Path $releaseDir "a-share-investment-workbench-portable-windows-amd64.zip"
Remove-Item -LiteralPath $archive -Force -ErrorAction SilentlyContinue
Compress-Archive -LiteralPath $portableRoot -DestinationPath $archive -CompressionLevel Optimal

Write-Host "Portable folder: $portableRoot"
Write-Host "Portable ZIP: $archive"
