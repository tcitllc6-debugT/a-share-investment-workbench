# Provider Boundaries

Provider packages isolate unstable external services from the Wails API and Vue UI.

## Rules

- New original-author service calls are not allowed.
- User-configured services must be optional.
- Public data endpoints must timeout and return typed empty or error states.
- UI code should not know concrete endpoint URLs.
- Provider tests should cover missing config and failure fallback.

## Current Providers

- `share`: exports analysis and shared text to local Markdown files.
- `update`: disables update checks by default until a replica-owned channel exists.
- `prompt`: uses the local prompt template library unless the user configures a remote endpoint.

## Planned Providers

- `market`: quote, K-line, minute line, market statistics.
- `news`: market news, telegraph, hot topics, calendar.
- `research`: announcements and research reports.
- `fund`: fund ranking, holdings, net value, fund K-line.

