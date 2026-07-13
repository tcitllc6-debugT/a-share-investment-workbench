import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const workflow = readFileSync(new URL('../../.github/workflows/a-share-update.yml', import.meta.url), 'utf8')
const macBuild = readFileSync(new URL('../../scripts/build-macos.sh', import.meta.url), 'utf8')
const windowsBuild = readFileSync(new URL('../../scripts/build-windows.ps1', import.meta.url), 'utf8')

test('GitHub release publishes the portable Windows package', () => {
  assert.match(workflow, /build-portable\.ps1/)
  assert.match(workflow, /a-share-investment-workbench-portable-windows-amd64\.zip/)
})

test('Windows release uses the reproducible pure-Go build mode', () => {
  assert.match(workflow, /CGO_ENABLED:\s*0/)
  assert.match(workflow, /go version -m/)
  assert.match(workflow, /\$buildInfoText = \[string\]::Join/)
})

test('the local Windows build script uses the reproducible pure-Go build mode', () => {
  assert.match(windowsBuild, /\$env:CGO_ENABLED\s*=\s*"0"/)
  assert.match(windowsBuild, /\$buildInfoText = \[string\]::Join/)
})

test('Windows release builds and verifies the frontend before embedding it', () => {
  assert.match(workflow, /- name: Build frontend bundle[\s\S]*working-directory: frontend[\s\S]*npm run build/)
  assert.match(workflow, /Test-Path -LiteralPath 'frontend\\dist\\index\.html'/)
  assert.match(windowsBuild, /Test-Path -LiteralPath \(Join-Path \$root "frontend\\dist\\index\.html"\)/)
})

test('GitHub release publishes a macOS DMG with the release version', () => {
  assert.match(workflow, /publish-macos-release:/)
  assert.match(workflow, /softprops\/action-gh-release@v2/)
  assert.match(workflow, /a-share-investment-workbench_\*_macos_universal\.dmg/)
  assert.match(macBuild, /APP_VERSION/)
  assert.match(macBuild, /max-old-space-size=6144/)
  assert.match(macBuild, /wails build --clean -s/)
})
