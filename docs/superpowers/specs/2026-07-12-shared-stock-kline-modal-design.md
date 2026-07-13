# Shared Stock K-line Modal Design

## Goal

Make the overview K-line action open the same in-app multi-period K-line experience used by pattern screening, without navigating away from the overview.

## Design

- Add a focused `StockKlineModal.vue` wrapper around `StockLightweightKlineChart.vue`.
- The wrapper owns the Naive UI modal layout, responsive width, scroll bounds, title, and chart mounting lifecycle.
- Callers own stock selection and code normalization, then bind `show`, `code`, `stockName`, and `darkTheme`.
- Replace the local modal markup in `SelectStock.vue` with the shared wrapper.
- Replace the overview event-and-route navigation with local modal state using the same wrapper.
- Preserve all multi-period chart controls and indicators because both entry points render the existing `StockLightweightKlineChart.vue` implementation.

## Error Handling

Invalid or empty stock codes do not open the modal. Closing the modal unmounts the chart so timers, subscriptions, and chart instances are released.

## Verification

- A source-contract regression test proves the overview no longer routes to `klineAnalysis` and both pages use `StockKlineModal`.
- Existing frontend tests verify that the surrounding workspaces remain intact.
- A production frontend build verifies Vue template compilation.
