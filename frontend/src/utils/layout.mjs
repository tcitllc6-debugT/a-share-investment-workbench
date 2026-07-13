export function getLayoutMode(viewportWidth) {
  if (viewportWidth < 760) return 'compact'
  if (viewportWidth < 1280) return 'standard'
  return 'wide'
}
