# Repository rules

## Web/mobile parity

The Flutter app in `mobile/` must stay in step with the web app. A previous
mobile app was deleted because it diverged too far from web to maintain.

A web feature change is done only when mobile has the equivalent, or
`docs/PARITY.md` records that mobile intentionally differs or has it planned.

A "web feature change" is any change to:

- pages under `web/app` (add, remove, rename a `page.tsx`),
- API clients under `web/lib/api`,
- gateway routes in `configs/envoy/envoy.yaml`.

How to satisfy the rule:

- Add or update the mobile screen/endpoint and flip the row in
  `docs/PARITY.md` to `done` (mobile screen and tests filled in), or
- add/adjust a `planned` row or a `differs: <reason>` row, or list the
  endpoint under "Planned endpoints" / "Intentionally differs" in
  `docs/PARITY.md` (format described at the top of that file).

App-wide UI choices where mobile deliberately looks or works unlike web
(navigation, bars hiding on scroll, player controls, the People tab) are
listed under "Mobile UI differences" in `docs/PARITY.md`. Keep them when
porting web changes, and update that list whenever one changes.

CI enforces this with `.github/workflows/parity.yml`, which runs
`cd mobile && dart run tool/parity_check.dart`. Run it locally before pushing
web, gateway, or mobile changes.

## Before changing anything in `mobile/`

Read these first, every time, so the change matches web and does not undo a
decision:

1. `docs/PARITY.md`: the row for the web page you are touching, and the
   "Mobile UI differences" list (deliberate choices to keep).
2. `docs/MOBILE_ISSUES.md`: known gaps. Your change may fix one (delete its
   entry) or must not make one worse.
3. `mobile/README.md`: how the app is built (shell and tabs, bars, players,
   l10n, push, release checklist).
4. The web counterpart under `web/app` and `web/components`: copy its
   behavior, fields, states and strings unless PARITY.md says otherwise.

While making the change:

- Strings come from the web locales (`web/lib/i18n/locales`). Add new ones
  there, then run `dart run tool/sync_locales.dart` in `mobile/`; never edit
  `mobile/l10n/*.arb` by hand.
- A change that makes mobile look or work unlike web goes into PARITY.md's
  "Mobile UI differences" (or the page's row) in the same commit.

Before you finish, run these in `mobile/`, then update the docs above for
anything that changed:
- `flutter analyze`
- `dart format --set-exit-if-changed lib test`
- `flutter test`
- `dart run tool/parity_check.dart --release`
- `dart run tool/sync_locales.dart --check`

Claude Code has the same checklist as the project skill `mobile-change`
(`.claude/skills/mobile-change/SKILL.md`).
