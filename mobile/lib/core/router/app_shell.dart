import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/router/app_tabs.dart';

/// The signed-in app frame: the current tab above a bottom navigation bar.
///
/// - The bar lists [tabs] (see `appTabs` for how a task adds one) and is
///   hidden while there are fewer than two (Flutter requires at least two
///   destinations).
/// - Each tab keeps its own navigation stack and state while another tab is
///   shown ([StatefulShellRoute.indexedStack]).
/// - The realtime socket pauses while the app is in the background and
///   resumes when it returns, so it does not hold a connection the OS will
///   kill anyway.
class AppShell extends ConsumerStatefulWidget {
  const AppShell({
    required this.navigationShell,
    required this.tabs,
    super.key,
  });

  /// go_router's handle on the tab branches: the current tab's screen and
  /// the tab switching.
  final StatefulNavigationShell navigationShell;

  /// The tabs to show, normally the registry [appTabs].
  final List<AppTab> tabs;

  /// The router branches for [tabs], one per tab.
  static StatefulShellRoute route(List<AppTab> tabs) =>
      StatefulShellRoute.indexedStack(
        builder: (context, state, shell) =>
            AppShell(navigationShell: shell, tabs: tabs),
        branches: [
          for (final tab in tabs)
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: tab.path,
                  builder: (context, state) => tab.builder(context),
                ),
              ],
            ),
        ],
      );

  @override
  ConsumerState<AppShell> createState() => _AppShellState();
}

class _AppShellState extends ConsumerState<AppShell> {
  late final AppLifecycleListener _lifecycle;

  @override
  void initState() {
    super.initState();
    _lifecycle = AppLifecycleListener(
      onPause: () => ref.read(realtimeManagerProvider).pause(),
      onResume: () => ref.read(realtimeManagerProvider).resume(),
    );
  }

  @override
  void dispose() {
    _lifecycle.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final tabs = widget.tabs;
    final l10n = AppLocalizations.of(context);
    final shell = widget.navigationShell;
    // Watched here (not in the destination) so the providers stay alive.
    final badges = [
      for (final tab in tabs) tab.badge == null ? 0 : ref.watch(tab.badge!),
    ];
    return Scaffold(
      body: shell,
      bottomNavigationBar: tabs.length < 2
          ? null
          : NavigationBar(
              selectedIndex: shell.currentIndex,
              // Tapping the current tab goes back to its root screen.
              onDestinationSelected: (index) => shell.goBranch(
                index,
                initialLocation: index == shell.currentIndex,
              ),
              destinations: [
                for (var i = 0; i < tabs.length; i++)
                  NavigationDestination(
                    icon: Badge.count(
                      count: badges[i],
                      isLabelVisible: badges[i] > 0,
                      child: Icon(tabs[i].icon),
                    ),
                    label: tabs[i].label(l10n),
                  ),
              ],
            ),
    );
  }
}
