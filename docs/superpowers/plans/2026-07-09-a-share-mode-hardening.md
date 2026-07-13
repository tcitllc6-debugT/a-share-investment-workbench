# A-Share Desk Replica Hardening Plan

## Goal

Finish the next replica-hardening pass in the isolated desktop copy and make A-share mode the product default.

## Tasks

1. Establish isolated working baseline.
   - Initialize git in the desktop replica source folder.
   - Commit the imported snapshot.
   - Work on `codex/a-share-mode-hardening`.

2. Add A-share runtime mode with tests.
   - Add `MarketFocusMode` to settings.
   - Default missing/unknown values to `a-share`.
   - Persist explicit overrides.
   - Test default and override behavior.

3. Clean high-impact replica residues.
   - Replace the embedded upstream manual with an A-Share Desk local manual.
   - Rename visible bot/tool/server labels from go-stock to A-Share Desk.
   - Remove remaining visible sponsor/VIP prompts from normal local flows where possible.

4. Refresh generated/static surfaces.
   - Regenerate Wails bindings if needed.
   - Rebuild frontend and assistant static assets.

5. Verify and package.
   - Run focused Go tests first, then broader build checks.
   - Build the Windows executable.
   - Copy the executable to the isolated `release` folder.
   - Launch from `release` with a clean working directory probe.

## Known Remaining Gaps To Report

- Market data still depends on public/third-party data providers.
- The UI remains the inherited Vue/Wails architecture rather than a full new design system.
- HK/US/fund modules are still present as optional modules unless explicitly removed later.
- Some upstream references remain in license/history/readme-style files for attribution.
