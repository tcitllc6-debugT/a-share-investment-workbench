# Replica Smoke Test

## Automated Build Verification

- [x] `npm --prefix frontend run build` passed on 2026-07-09.
- [x] `go test ./... -run '^$' -count=0` passed on 2026-07-09.
- [x] `wails build` passed on 2026-07-09 and produced `build/bin/a-share-desk.exe`.
- [x] Launch probe started `build/bin/a-share-desk.exe` and confirmed the process stayed alive for 8 seconds.

## Launch

- [ ] App opens without requiring original-author service access.
- [ ] SQLite database initializes under `data/`.
- [ ] Window title and About page show `A-Share Desk`.
- [ ] Update check reports the local disabled update channel unless a replica-owned provider is added.

## Core Workflows

- [ ] Add a stock to watchlist.
- [ ] Create a group.
- [ ] Refresh price.
- [ ] Open K-line view.
- [ ] Open market/news tab.
- [ ] Configure AI provider.
- [ ] Run AI analysis.
- [ ] Export or share analysis locally.
- [ ] Open settings.
- [ ] Open About page.

## Negative Cases

- [ ] App handles missing AI key.
- [ ] App handles network failure.
- [ ] App handles empty market data.
- [ ] App handles disabled update provider.

## External-Service Guardrails

- [ ] Prompt plaza stays local when no API base is configured.
- [ ] Sharing does not upload to original-author services.
- [ ] Local-only features do not require sponsor, VIP, or device binding.
- [ ] UI code does not add new hardcoded original-author service calls.
