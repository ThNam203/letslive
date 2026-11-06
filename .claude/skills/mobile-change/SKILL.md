---
name: mobile-change
description: Use before changing anything in the Flutter app under mobile/ (screens, widgets, providers, strings, tests), and when porting a web change to mobile, so the change matches the web app and keeps the documented mobile decisions.
---

# Changing the mobile app

The Flutter app in `mobile/` must match the Next.js app in `web/`, except
where the docs record a deliberate difference. A previous mobile app was
deleted because it drifted from web, so read before you write.

## 1. Read first

- `AGENTS.md`, sections "Web/mobile parity" and "Before changing anything in
  `mobile/`": the rule and the checklist.
- `docs/PARITY.md`:
  - the row for the web page you are touching;
  - the "Mobile UI differences" list (keep those choices).
- `docs/MOBILE_ISSUES.md`: known gaps, which your change may fix or touch.
- `mobile/README.md`: the section for the area (shell and tabs, players,
  notifications, l10n, release checklist).
- The web counterpart (`web/app/...`, `web/components/...`,
  `web/lib/api/...`): its fields, states, errors and strings.

## 2. While changing

- Match web unless PARITY.md says mobile differs. If the request makes mobile
  differ, record it in "Mobile UI differences" or the page's row.
- New strings go into `web/lib/i18n/locales/{en-US,vi-VN}/*.json`, then
  `dart run tool/sync_locales.dart` in `mobile/`. Never edit the ARB files
  by hand.
- Tab root screens use `TabScaffold`; watch screens use `WatchScaffold`; the
  players share `PlayerChrome`. Reuse these instead of new frames.
- Keep the code simple and documented, with tests for new behavior.

## 3. Before finishing

Run in `mobile/`:

```bash
flutter analyze
dart format --set-exit-if-changed lib test
flutter test
dart run tool/parity_check.dart --release
dart run tool/sync_locales.dart --check
```

Then update the docs in the same commit:
- `docs/PARITY.md` (row and differences);
- `docs/MOBILE_ISSUES.md` (delete fixed entries, add new gaps);
- `mobile/README.md` when how the app is built changed.
