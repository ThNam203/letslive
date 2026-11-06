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
