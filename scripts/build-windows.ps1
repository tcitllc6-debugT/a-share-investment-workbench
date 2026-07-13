param(
  [string]$Version = "1.2.0"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
$env:CGO_ENABLED = "0"

function Invoke-CodeSign {
  param([string]$Path)
  if (-not $env:WINDOWS_SIGN_CERT_THUMBPRINT) {
    Write-Warning "WINDOWS_SIGN_CERT_THUMBPRINT 未配置，产物将是未签名版本。"
    return
  }
  $signtool = Get-Command signtool.exe -ErrorAction SilentlyContinue
  if (-not $signtool) {
    throw "配置了 WINDOWS_SIGN_CERT_THUMBPRINT 但未找到 signtool.exe"
  }
  $timestamp = if ($env:WINDOWS_TIMESTAMP_URL) { $env:WINDOWS_TIMESTAMP_URL } else { "http://timestamp.digicert.com" }
  & $signtool.Source sign /sha1 $env:WINDOWS_SIGN_CERT_THUMBPRINT /fd SHA256 /tr $timestamp /td SHA256 $Path
}

Write-Host "构建前端..."
Push-Location frontend
npm run build
Pop-Location
if (-not (Test-Path -LiteralPath (Join-Path $root "frontend\dist\index.html"))) {
  throw "前端构建产物 frontend/dist/index.html 不存在，终止打包以避免生成无界面程序"
}

Write-Host "构建 A 股智投工作台主程序并生成 Wails 绑定..."
wails build -clean -s -ldflags "-X main.Version=$Version"

Write-Host "构建热更新助手..."
go build -trimpath -ldflags "-s -w" -o build\bin\a-share-updater.exe .\backend\cmd\a-share-updater
$application = Get-ChildItem (Join-Path $root "build\bin") -Filter "*.exe" -File |
  Where-Object { $_.Name -ne "a-share-updater.exe" -and $_.Name -notlike "*-installer.exe" } |
  Sort-Object LastWriteTime -Descending |
  Select-Object -First 1
if (-not $application) { throw "Wails 主程序构建产物不存在" }
$buildInfo = & go version -m $application.FullName
$buildInfoText = [string]::Join([Environment]::NewLine, [string[]]$buildInfo)
if ($buildInfoText -notmatch 'CGO_ENABLED=0') { throw "Windows 发布必须使用 CGO_ENABLED=0，实际产物：$($application.FullName)" }
Invoke-CodeSign $application.FullName
Invoke-CodeSign (Join-Path $root "build\bin\a-share-updater.exe")

Write-Host "生成 NSIS 安装包..."
$makensis = Get-Command makensis.exe -ErrorAction SilentlyContinue
if (-not $makensis) {
  $fallbackNsis = (Join-Path ([Environment]::GetFolderPath('ProgramFilesX86')) "NSIS\makensis.exe")
  if (Test-Path $fallbackNsis) { $makensis = Get-Item $fallbackNsis }
}
if (-not $makensis) { throw "未找到 makensis.exe" }
$makensisPath = $makensis.Source
if (-not $makensisPath) { $makensisPath = $makensis.FullName }
& $makensisPath "-DARG_WAILS_AMD64_BINARY=$($application.FullName)" "$root\build\windows\installer\project.nsi"
$installer = Get-ChildItem build\bin\*-installer.exe | Sort-Object LastWriteTime -Descending | Select-Object -First 1
if (-not $installer) { throw "未找到 NSIS 安装包" }
Invoke-CodeSign $installer.FullName
Write-Host "安装包：$($installer.FullName)"
