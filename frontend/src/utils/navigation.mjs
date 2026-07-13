export function getNavigationPresentation(viewportWidth) {
  if (viewportWidth < 960) return 'drawer'
  if (viewportWidth < 1280) return 'standard'
  return 'wide'
}

export function getTableScrollWidth(kind, viewportWidth) {
  const minimums = { radar: 940, screening: 860, watchlist: 980 }
  const minimum = minimums[kind] || 760
  return viewportWidth < 960 ? minimum : Math.max(minimum, viewportWidth - 260)
}
