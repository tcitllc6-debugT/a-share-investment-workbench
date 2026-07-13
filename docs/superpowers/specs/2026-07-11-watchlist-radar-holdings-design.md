# Watchlist, Radar, and Holdings Design

## Scope

This change fixes the watchlist display regression, makes the intraday radar
understandable, keeps holding metrics readable at every desktop scale, and
defines a first release of the AI research committee report. It also produces
a standalone desktop project folder without changing the upstream repository.

## Watchlist

Watchlist membership and realtime quote payloads currently use several code
formats, such as `sz300418`, `300418.SZ`, and `SZ300418`. The card renderer
must compare a canonical identity, not raw strings. A small pure frontend
normalizer will normalize A-share, Beijing, Hong Kong, and US symbols before
membership checks. After a successful follow action, the current group will be
reloaded, its symbols synchronized, and a quote refresh requested. A newly
followed symbol remains visible while its first quote is loading.

## Intraday Radar

The radar remains discovery, notification, and analysis navigation only. It
does not send orders. It scans market change events every 15 seconds while
enabled and groups them as:

- Opportunity: momentum, strength, and limit-up events.
- Risk: risk and breakdown events.
- Capital: large-order/capital-flow events.

Each event will display a plain-language trigger explanation, priority, event
time, data category, and a suggested next action. The page will include a
compact "How this radar works" disclosure explaining P1/P2, the 15-second
refresh, empty-state meaning, and that an event is a cue for review rather
than investment advice.

## Holdings

The workspace label becomes "Stock Holdings". The six summary metrics remain
on one horizontal strip at desktop widths. At narrow widths the strip has
horizontal overflow rather than wrapping into a misleading or cramped grid.
The trading-record ledger remains below it as the sole accounting source.

## Window Scrolling

The workspace main region owns vertical scrolling. Context navigation keeps
its own vertical scroll. Wide tables and the holdings summary provide explicit
horizontal scrolling, allowing a small or high-DPI window to reach every
control without clipping the application surface.

## AI Research Committee Report

Phase one is a manual, single-stock report entered from watchlist, K-line, or
radar analysis. It must make sources and uncertainty visible rather than emit
a bare recommendation. Its sections are technicals, fundamentals, policy and
industry, corporate events (announcements/reductions/unlocks), sentiment and
Dragon Tiger activity, evidence, conflicts, and conditional conclusions.
The report is a design contract in this scope; report-generation APIs and
scheduled batch generation are deliberately deferred until source reliability
and model-cost controls are specified.

## Independent Delivery Directory

The current working source remains at its existing independent desktop path.
After verification, a clean desktop folder named `A股智投工作台` will contain
`source` and `release`. It will be copied without `node_modules`, generated
frontend bindings, local databases, logs, or user portfolio data.

