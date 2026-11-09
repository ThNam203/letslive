# Mobile issues — letslive

_Last updated: 2026-11-06_

Known gaps in the Flutter app (`mobile/`): tooling, structure, behavior that
works but is not finished, and checks never run on a device. Deliberate
differences from web are not issues; they are listed under "Mobile UI
differences" in [PARITY.md](PARITY.md).

When you fix one, delete its entry in the same commit. When you find one,
add it here with a fix and the files involved.

---

## 🟠 HIGH

**M5. The profile header does not scroll away**
On the Profile tab and on channels, only the tab's top bar hides while
scrolling. The banner, avatar block and tab strip stay pinned, and just the
tab content below them scrolls, so most of the screen never gets freed. On
YouTube the whole channel header scrolls away.
Fix: build `ProfileTabs` on a `NestedScrollView` (header and `TabBar` in
`headerSliverBuilder`, the tab bar pinned), and keep the live layout's chat
sizing working on channels.
Files: [mobile/lib/features/profile/presentation/profile_tabs.dart](../mobile/lib/features/profile/presentation/profile_tabs.dart), [mobile/lib/features/watch/presentation/channel_screen.dart](../mobile/lib/features/watch/presentation/channel_screen.dart)

**M6. `live_player_view.dart` is a grab-bag shared by both players**
Besides the live player it holds what the VOD player also uses:
- `PlayPauseButton`, `PlayerSettingsMenu`, `PlayerSpinner` and `PlayerError`;
- `isLandscape` and `LandscapePlayer`;
- `WatchScaffold`, the frame of the watch screens.

It and `player_chrome.dart` import each other.
Fix: move the shared pieces into one shared player file (e.g.
`player_chrome.dart`), and the watch-screen frame next to it. Keep
`live_player_view.dart` for the live player only. No behavior change.
Files: [mobile/lib/features/watch/presentation/live_player_view.dart](../mobile/lib/features/watch/presentation/live_player_view.dart), [mobile/lib/features/watch/presentation/player_chrome.dart](../mobile/lib/features/watch/presentation/player_chrome.dart), [mobile/lib/features/vod/presentation/vod_player_view.dart](../mobile/lib/features/vod/presentation/vod_player_view.dart)

**M7. No device check of the 2026-11-06 UI changes**
None of these have been seen on a real phone:
- the bars hiding on scroll (animation, safe areas, status bar icons);
- tap to pause, and the controls fading;
- full screen in and out, and picture-in-picture from the gear menu;
- the People tab.

Only the full-width video description was confirmed on a phone.
Fix: go through the README "Release checklist" rotation and PiP items plus
these, on Android and iOS.
File: [mobile/README.md](../mobile/README.md)

## 🟡 MEDIUM

**M1. `parity_check` accounting is coarse**
A prefix such as `/user` or `/vods` accounts for every sub-endpoint under it.
`fetchClient` calls with non-literal paths are skipped. Changes under
`web/types` are not checked, although the spec asks for it.
Fix: per-path accounting, and a warning on skipped calls.
File: [mobile/tool/parity_check.dart](../mobile/tool/parity_check.dart)

**M8. The system bars may change after full screen**
`LandscapePlayer` hides the status and navigation bars, and on dispose sets
`SystemUiMode.edgeToEdge`. Nothing in the app set a mode before, so on older
Android versions the bars may turn transparent for the rest of the session
after the first full screen. Android 15+ is edge-to-edge anyway.
Fix: set `edgeToEdge` once at startup (`main.dart`), so the app looks the
same before and after full screen, or restore the previous mode instead.
File: [mobile/lib/features/watch/presentation/live_player_view.dart](../mobile/lib/features/watch/presentation/live_player_view.dart)

**M9. Leaving full screen with the button locks the screen upright**
The exit button locks portrait until the watch screen closes. Otherwise a
phone still held sideways would turn straight back. Until then, turning the
phone no longer enters full screen on that screen (the button still does).
Fix: unlock once the phone is physically upright again. That needs the
device orientation from a sensor package, which is a new dependency (ask the
owner first).
File: [mobile/lib/features/watch/presentation/player_chrome.dart](../mobile/lib/features/watch/presentation/player_chrome.dart)

**M10. Missing tests for the 2026-11-06 changes**
- The bars hiding is tested only in a stand-in tab
  (`test/core/router/app_shell_test.dart`), not on the real Home, Messages
  and Profile screens (pull-to-refresh, nested tab content).
- Nothing checks that leaving a watch screen calls `resetPlayerOrientation`.
- Nothing checks the immersive mode that `LandscapePlayer` sets and restores.

Fix: add those tests, recording `SystemChannels.platform` calls as
`test/features/watch/player_chrome_test.dart` does.
Files: [mobile/test/core/router/app_shell_test.dart](../mobile/test/core/router/app_shell_test.dart), [mobile/test/features/vod/vod_screen_test.dart](../mobile/test/features/vod/vod_screen_test.dart), [mobile/test/features/watch/channel_screen_test.dart](../mobile/test/features/watch/channel_screen_test.dart)

## 🟢 LOW

**M11. The bars slide on a timer instead of following the finger**
The tab's top bar and the bottom bar animate in and out over 200 ms when the
scroll direction changes, and the list jumps by the bar's height. YouTube's
bar moves with the scroll.
Fix: a floating, snapping `SliverAppBar` per tab, with each tab's list as
slivers. Do it only if the current slide feels wrong on a device.
Files: [mobile/lib/core/router/app_chrome.dart](../mobile/lib/core/router/app_chrome.dart), the tab screens

**M12. The hidden bottom bar leaves lists under the gesture area and hides the badge**
With the bottom bar away the last list row can sit under the system gesture
bar, and the unread-messages badge is hidden too.
Fix: keep the bottom safe-area padding in the tab body while the bar is
hidden; decide with the owner whether the badge should stay visible.
File: [mobile/lib/core/router/app_shell.dart](../mobile/lib/core/router/app_shell.dart)

**M2. Plural forms are not reconciled across locales**
`sync_locales` copies each plural key as it is; it does not check that `en`
and `vi` have the same categories.
Fix: check that each plural key has the same categories in both locales.
File: [mobile/tool/sync_locales.dart](../mobile/tool/sync_locales.dart)

**M3. `Info-Dev.plist` duplicates `Info.plist`**
Every plist change (purpose strings, URL types) must be made in both files.
Fix: one plist, with the dev ATS exception from a build setting.
Files: [mobile/ios/Runner/Info.plist](../mobile/ios/Runner/Info.plist), [mobile/ios/Runner/Info-Dev.plist](../mobile/ios/Runner/Info-Dev.plist)

**M4. Feed cards refetch their owner on scroll-back**
`publicUserProvider` is an `autoDispose` family, so a card scrolled out and
back in fetches its owner again.
Fix: `ref.keepAlive()` with a short timer, or pass the owner from the list
response.
File: [mobile/lib/features/user/user_providers.dart:26-30](../mobile/lib/features/user/user_providers.dart#L26-L30)
