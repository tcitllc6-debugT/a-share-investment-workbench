export function buildRadarFilter({tab = 'all', onlyFollowed = false, onlyHoldings = false, minAmount = 0}) {
  const categories = tab === 'risk' ? ['risk', 'breakdown'] : tab === 'opportunity' ? ['momentum', 'strength', 'limit'] : tab === 'capital' ? ['capital'] : []
  return {categories, onlyFollowed, onlyHoldings, minAmount}
}
