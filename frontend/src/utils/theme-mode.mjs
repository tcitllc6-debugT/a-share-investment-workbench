export const THEME_MODE_STORAGE_KEY = 'a-share-theme-mode'

const THEME_MODES = new Set(['light', 'dark'])

export function resolveThemeMode(preferredMode, storedMode) {
  if (THEME_MODES.has(storedMode)) {
    return storedMode
  }
  return preferredMode === 'dark' ? 'dark' : 'light'
}

export function persistThemeMode(mode, storage = globalThis.localStorage) {
  const resolved = resolveThemeMode(mode)
  storage?.setItem?.(THEME_MODE_STORAGE_KEY, resolved)
  return resolved
}

export function applyThemeMode(mode, documentRef = globalThis.document) {
  const resolved = resolveThemeMode(mode)
  const root = documentRef?.documentElement

  if (root) {
    root.dataset.themeMode = resolved
    root.setAttribute('data-theme-mode', resolved)
  }

  return resolved
}
