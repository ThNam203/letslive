# Web / mobile parity

The Flutter app in `mobile/` must keep the web app's feature set. This file is
the ledger; `cd mobile && dart run tool/parity_check.dart` (CI:
`.github/workflows/parity.yml`) enforces it. See `AGENTS.md` for the rule.

Source of the feature list: section 4 of
`docs/superpowers/specs/2026-11-04-flutter-mobile-app-design.md`.

## How to edit this file

- **Screens table**: one row per web page (`page.tsx` under `web/app`, route
  groups like `(main)` removed, dynamic segments such as `[userId]` kept).
  Columns: `web route/feature | mobile screen | tests | status`.
  - Status is exactly one of `done`, `planned`, or `differs: <reason>`.
  - To flip a row to `done`: put the mobile screen/widget in column 2 and the
    test file(s) in column 3 (both required), then set status `done`.
  - Rows whose first cell starts with `/` must match a real web page; other
    first cells are free-form feature rows.
- **Planned endpoints** and **Intentionally differs** are bullet lists:
  ``- `/path` - reason``. An entry accounts for every gateway/web-client path
  that starts with it (whole segments). The reason is required.
  - When a task adds an `ApiEndpoints` constant covering a planned entry,
    delete that bullet in the same commit (the constant now accounts for it).
- `dart run tool/parity_check.dart --release` additionally fails while any
  row is `planned` or any "Planned endpoints" bullet remains (Task 20).

## Screens

| web route/feature | mobile screen | tests | status |
|---|---|---|---|
| `/` | `HomeScreen` + `AppShell` (`mobile/lib/features/home/presentation/home_screen.dart`, `mobile/lib/core/router/app_shell.dart`): live, followed/recommended channels and popular videos | `mobile/test/features/home/home_screen_test.dart`, `mobile/test/features/home/home_repository_test.dart`, `mobile/test/core/router/app_shell_test.dart`, `mobile/contract_test/home_contract_test.dart` | done |
| `/login` | `LoginScreen` (`mobile/lib/features/auth/presentation/login_screen.dart`) | `mobile/test/features/auth/login_screen_test.dart`, `mobile/contract_test/auth_contract_test.dart`, `mobile/integration_test/auth_flow_test.dart` | done |
| `/signup` | `SignupScreen` + `OtpDialog` (`mobile/lib/features/auth/presentation/`) | `mobile/test/features/auth/signup_screen_test.dart`, `mobile/test/features/auth/auth_validators_test.dart`, `mobile/contract_test/auth_contract_test.dart` | done |
| `/account-setup` | `AccountSetupScreen` (`mobile/lib/features/auth/presentation/account_setup_screen.dart`) | `mobile/test/features/auth/account_setup_screen_test.dart`, `mobile/test/core/router/app_router_test.dart`, `mobile/contract_test/auth_contract_test.dart` | done |
| `/cookie-policy` | - | - | planned |
| `/search` | `SearchScreen` (`mobile/lib/features/search/presentation/search_screen.dart`): debounced box, tabs All / Channels / Live / Videos, accent-insensitive highlight | `mobile/test/features/search/search_screen_test.dart`, `mobile/test/features/search/search_repository_test.dart`, `mobile/test/features/search/highlight_test.dart`, `mobile/contract_test/search_contract_test.dart` | done |
| `/messages` | - | - | planned |
| `/messages/[conversationId]` | - | - | planned |
| `/notifications` | - | - | planned |
| `/settings/chat-commands` | - | - | planned |
| `/settings/profile` | - | - | planned |
| `/settings/security` | - | - | planned |
| `/settings/stream` | - | - | planned |
| `/settings/upload` | - | - | planned |
| `/settings/vods` | - | - | planned |
| `/shop` | - | - | planned |
| `/users/[userId]` | `ChannelScreen` (`mobile/lib/features/watch/presentation/channel_screen.dart`): live player, offline / ended / not-found states (Task 13a); profile header, tabs, follow and videos come in Task 15 | `mobile/test/features/watch/channel_screen_test.dart`, `mobile/test/features/watch/channel_navigation_test.dart`, `mobile/contract_test/watch_contract_test.dart` | planned |
| `/users/[userId]/gifts` | - | - | planned |
| `/users/[userId]/vods/[vodId]` | - | - | planned |
| `/wallet/deposit` | - | - | differs: out of scope (Stripe deposits/payments excluded) |
| `/wallet/inventory` | - | - | planned |
| `/wallet/overview` | - | - | planned |
| `/wallet/transactions` | - | - | planned |

## Planned endpoints

Gateway / web-client paths mobile will call but has no `ApiEndpoints`
constant for yet. Remove a bullet when its constant lands.

- `/auth/password` - update password (settings/security)
- `/user` - public profile, follow/unfollow, own profile, stream key, notifications, inventory, gifts
- `/upload-file` - profile/background picture upload
- `/currencies` - wallet
- `/wallet` - wallet overview
- `/transactions` - wallet transactions
- `/shop` - shop items and purchase
- `/vods` - VOD feed, detail, reactions, comments, upload
- `/vod-comments` - comment edit, delete, like, replies
- `/messages` - live chat history and send
- `/chat-commands` - chat command settings
- `/conversations` - direct messages
- `/files` - media host path (thumbnails, VOD segments) used by web image/player URLs

## Intentionally differs

Gateway prefixes mobile never calls through `ApiEndpoints`.

- `/realtime` - WebSocket, not REST: mobile uses `AppConfig.wsBaseUrl` (`<ws origin>/realtime`)
- `/deposits` - Stripe deposits are out of scope for mobile
- `/payments` - Stripe payment history is out of scope for mobile
- `/gifts` - gateway route with no backend handler (user service only has `/user/{id}/gifts/received`, `/user/me/gifts/sent`) and no web caller
- `/livestreamings` - gateway route with no backend handler and no web caller
- `/is-streaming` - gateway route with no backend handler and no web caller
