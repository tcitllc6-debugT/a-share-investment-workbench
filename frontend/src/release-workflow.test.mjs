import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const workflow = readFileSync(new URL('../../.github/workflows/a-share-update.yml', import.meta.url), 'utf8')
const macBuild = readFileSync(new URL('../../scripts/build-macos.sh', import.meta.url), 'utf8')

test('GitHub release publishes the portable Windows package', () => {
  assert.match(workflow, /build-portable\.ps1/)
  assert.match(workflow, /A股智投工作台-便携版\.zip/)
})

test('GitHub release publishes a macOS DMG with the release version', () => {
  assert.match(workflow, /publish-macos-release:/)
  assert.match(workflow, /softprops\/action-gh-release@v2/)
  assert.match(workflow, /a-share-investment-workbench_\*_macos_universal\.dmg/)
  assert.match(macBuild, /APP_VERSION/)
})
