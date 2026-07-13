# A-Share Desk Replica Hardening Design

## Intent

This pass keeps the project as an isolated desktop-first replica and sharpens it for A-share users. It must not depend on the original hosted services, hidden sponsor flows, or upstream release/update channels.

## Decisions

- Default market focus is `a-share`.
- HK, US, fund, MCP, and agent features remain available as optional local modules, but the default runtime should describe itself as A-share-first.
- Original sponsor/VIP checks stay as compatibility shims only; they should not block local features.
- Original hosted endpoints must stay absent from product runtime paths.
- User-facing brand should be A-Share Desk, except license/history attribution in docs.

## Scope

- Add a small runtime setting for market focus, defaulting to `a-share`.
- Add tests for the default mode and preserved user overrides.
- Replace remaining high-impact visible upstream labels in assistant/bot/tool descriptions.
- Replace the embedded user manual with an A-Share Desk local manual.
- Rebuild static assistant assets when tooling is available.

## Out of Scope

- Rewriting the full UI layout from scratch.
- Replacing third-party market data providers.
- Removing GPL attribution or historical upstream references from license-style documents.
