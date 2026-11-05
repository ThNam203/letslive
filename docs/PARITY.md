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
| `/cookie-policy` | - | - | differs: native app sets no browser cookies; policy linked from the store listing |
| `/search` | `SearchScreen` (`mobile/lib/features/search/presentation/search_screen.dart`): debounced box, tabs All / Channels / Live / Videos, accent-insensitive highlight | `mobile/test/features/search/search_screen_test.dart`, `mobile/test/features/search/search_repository_test.dart`, `mobile/test/features/search/highlight_test.dart`, `mobile/contract_test/search_contract_test.dart` | done |
| `/messages` | `ConversationsScreen` (`mobile/lib/features/messages/presentation/conversations_screen.dart`) as the Messages tab, with the unread badge on the tab icon; new direct message or group via `NewConversationScreen` (user search, group name, member limit). Rows bubble to the top on a new message (web updates in place). Web's chat dock and header popover are desktop-only, not ported; there is no mute, add/remove member, leave or rename UI because web has none (the repository covers those endpoints) | `mobile/test/features/messages/messages_screens_test.dart`, `mobile/test/features/messages/dm_state_test.dart`, `mobile/test/features/messages/dm_units_test.dart`, `mobile/contract_test/messages_contract_test.dart` | done |
| `/messages/[conversationId]` | `ConversationScreen` (`mobile/lib/features/messages/presentation/conversation_screen.dart`): header with online state, messages paged by cursor, typing indicator, emote picker and composer, read marking, reconnect catch-up, not-found state; long-press on an own message edits or deletes it (web's API has this but no UI). Sending pictures is not ported yet: it needs the general file upload (`/upload-file`, Task 19); received pictures are shown | `mobile/test/features/messages/messages_screens_test.dart`, `mobile/test/features/messages/dm_state_test.dart`, `mobile/contract_test/messages_contract_test.dart` | done |
| `/notifications` | `NotificationsScreen` (`mobile/lib/features/notifications/presentation/notifications_screen.dart`) as the Notifications tab (between Search and Messages) with the unread badge on the tab icon: list newest first and paged, unread dot, open what it is about (`resolveDeepLink`, only the app's own routes), mark read, mark all read, delete; new ones arrive live (`notification.created`) and a reconnect loads them again. Push notifications (`PushService`, `mobile/lib/core/push/`): the FCM token is registered for the signed-in user and removed on logout or session expiry, taps open the screen the push names; foreground pushes are not shown (the app is already connected live). Web's header bell popover is desktop-only, replaced by the tab; relative times are computed when a row is built, not ticked every minute; the mark-all action is an icon (a text button overflows a narrow bar) | `mobile/test/features/notifications/notifications_screen_test.dart`, `mobile/test/features/notifications/push_flow_test.dart`, `mobile/test/core/push/push_service_test.dart`, `mobile/test/core/push/deep_link_test.dart`, `mobile/contract_test/notifications_contract_test.dart` | done |
| `/settings/chat-commands` | - | - | planned |
| `/settings/profile` | `ProfileSettingsScreen` (`mobile/lib/features/settings/presentation/profile_settings_screen.dart`): banner and profile picture from the photo library with upload progress, username, bio and social links; reached from the Profile tab via `SettingsScreen`. Web's crop dialog, theme, language and disable-account sections are not ported (the server needs JPEG or WebP, so PNG and GIF are refused before upload) | `mobile/test/features/profile/profile_settings_test.dart`, `mobile/test/features/profile/profile_units_test.dart`, `mobile/contract_test/profile_contract_test.dart` | done |
| `/settings/security` | `SecurityScreen` (`mobile/lib/features/settings/presentation/security_screen.dart`): account email and the change-password form (web `changePassword.ts` rules, `PATCH /auth/password`). Phone number, stream key and two-factor are not ported | `mobile/test/features/profile/security_screen_test.dart`, `mobile/test/features/profile/profile_units_test.dart`, `mobile/contract_test/profile_contract_test.dart` | done |
| `/settings/stream` | - | - | planned |
| `/settings/upload` | - | - | planned |
| `/settings/vods` | - | - | planned |
| `/shop` | `ShopScreen` (`mobile/lib/features/shop/presentation/shop_screen.dart`): catalogue with price and a Buy button, reached from the Profile tab's app bar. One purchase request, never retried; every Buy button is disabled while it runs; balance and inventory reload afterwards, also after a failure. Errors show the translation of the server's error key (e.g. insufficient balance). Items are rows, not web's grid cards | `mobile/test/features/shop/shop_test.dart`, `mobile/test/features/wallet/wallet_test_app.dart`, `mobile/contract_test/wallet_contract_test.dart` | done |
| `/users/[userId]` | `ChannelScreen` (`mobile/lib/features/watch/presentation/channel_screen.dart`): live player, offline / ended / not-found states (Task 13a); live chat with emotes and slash commands (Task 13b, `mobile/lib/features/watch/chat/`, `mobile/lib/core/chat_parser/`); Tab/arrow suggestion navigation is desktop-only, not ported; profile header, follow button and the tabs Chat / Home / Videos / About (`mobile/lib/features/profile/`; chat is a tab because a phone cannot show it beside the profile). The Profile tab (`OwnProfileScreen`) shows the signed-in user's own profile. Gift button (`GiftButton`, `mobile/lib/features/shop/presentation/gift_button.dart`) next to the follow button in the profile header, opening the gift picker as a bottom sheet (`GiftSheet`): picking an item buys and sends it (one request, never retried; web sends no message). The video page's uploader row has it too, beside the follow button (`mobile/lib/features/vod/presentation/vod_details.dart`) | `mobile/test/features/watch/channel_screen_test.dart`, `mobile/test/features/watch/channel_navigation_test.dart`, `mobile/contract_test/watch_contract_test.dart`, `mobile/test/features/watch/chat/chat_panel_test.dart`, `mobile/test/features/watch/chat/chat_controller_test.dart`, `mobile/test/core/chat_parser/command_test.dart`, `mobile/contract_test/chat_contract_test.dart`, `mobile/test/features/profile/profile_tabs_test.dart`, `mobile/test/features/profile/follow_test.dart`, `mobile/contract_test/profile_contract_test.dart`, `mobile/test/features/shop/shop_test.dart`, `mobile/contract_test/wallet_contract_test.dart` | done |
| `/users/[userId]/gifts` | `GiftsScreen` (`mobile/lib/features/shop/presentation/gifts_screen.dart`): gifts a user received, paged (web shows only the first 20), opened from the About tab of the profile (web has no link to the page, only the URL). The backend's sent-gifts list (`GET /user/me/gifts/sent`) is in `GiftRepository.sent` and contract-tested; web has no screen for it, so neither has mobile | `mobile/test/features/shop/shop_test.dart`, `mobile/test/features/wallet/wallet_test_app.dart`, `mobile/contract_test/wallet_contract_test.dart` | done |
| `/users/[userId]/vods/[vodId]` | `VodScreen` (`mobile/lib/features/vod/presentation/vod_screen.dart`): player with seek bar, details, like/dislike, other streams, comments with replies, likes, edit and delete, tappable timestamps; a copy-link button (clipboard, needs `WEB_BASE_URL`; web's share dialog and its "start at current time" switch are not ported); the uploader row has the follow button (`FollowButton`, `mobile/test/features/profile/follow_test.dart`) | `mobile/test/features/vod/vod_screen_test.dart`, `mobile/test/features/vod/vod_comments_test.dart`, `mobile/test/features/vod/vod_reaction_test.dart`, `mobile/test/features/vod/timestamps_test.dart`, `mobile/test/features/vod/view_threshold_test.dart`, `mobile/test/features/watch/channel_navigation_test.dart`, `mobile/contract_test/vod_contract_test.dart` | done |
| `/wallet/deposit` | - | - | differs: deposits/payments are out of scope for the mobile release |
| `/wallet/inventory` | `InventoryTab` (`mobile/lib/features/shop/presentation/inventory_tab.dart`), the third tab of `WalletScreen`: owned items with quantity, paged (web shows the first 20), unknown items as "Unknown item" | `mobile/test/features/wallet/wallet_screen_test.dart`, `mobile/test/features/shop/shop_test.dart`, `mobile/contract_test/wallet_contract_test.dart` | done |
| `/wallet/overview` | `WalletScreen` + `WalletOverviewTab` (`mobile/lib/features/wallet/presentation/`), reached from the Profile tab's app bar: a card per currency (SPARK, FLARE; 0 when missing) and the five latest transactions. Amounts are exact decimals (`DecimalAmount`), never doubles, formatted as web does (grouping, at most 2 fraction digits). No deposit entry point | `mobile/test/features/wallet/wallet_screen_test.dart`, `mobile/test/features/wallet/decimal_amount_test.dart`, `mobile/contract_test/wallet_contract_test.dart` | done |
| `/wallet/transactions` | `WalletTransactionsTab` (`mobile/lib/features/wallet/presentation/transactions_tab.dart`): type filter chips (web's set, no deposit) over the loaded rows, next page loads as the list scrolls (web has a Load more button) | `mobile/test/features/wallet/wallet_screen_test.dart`, `mobile/contract_test/wallet_contract_test.dart` | done |

## Planned endpoints

Gateway / web-client paths mobile will call but has no `ApiEndpoints`
constant for yet. Remove a bullet when its constant lands.

- `/user` - public profile, follow/unfollow, own profile, stream key, inventory, gifts
- `/upload-file` - general file upload (settings/stream thumbnails); profile and banner pictures use `/user/me/*-picture` instead
- `/vods/upload` - VOD upload (settings/upload)
- `/vods/author` - the signed-in user's own VODs (settings/vods)
- `/vods/{vodId}` - PATCH/DELETE of an own VOD (settings/vods, Task 19); GET is built (video screen). Doc only: `ApiEndpoints.vods` already satisfies the checker
- `/files` - media host path (thumbnails, VOD segments) used by web image/player URLs

## Intentionally differs

Gateway prefixes mobile never calls through `ApiEndpoints`.

- `/realtime` - WebSocket, not REST: mobile uses `AppConfig.wsBaseUrl` (`<ws origin>/realtime`)
- `/deposits` - Stripe deposits are out of scope for mobile
- `/payments` - Stripe payment history is out of scope for mobile
- `/gifts` - gateway route with no backend handler (user service only has `/user/{id}/gifts/received`, `/user/me/gifts/sent`) and no web caller
- `/livestreamings` - gateway route with no backend handler and no web caller
- `/is-streaming` - gateway route with no backend handler and no web caller
