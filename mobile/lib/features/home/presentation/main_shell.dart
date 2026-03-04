import 'package:flutter/material.dart';
import 'package:forui/forui.dart';
import 'package:go_router/go_router.dart';

import '../../../core/router/app_router.dart';
import '../../../l10n/app_localizations.dart';

class MainShell extends StatelessWidget {
  final Widget child;

  const MainShell({super.key, required this.child});

  int _currentIndex(BuildContext context) {
    final location = GoRouterState.of(context).uri.path;
    if (location.startsWith(AppRoutes.messages)) return 1;
    if (location.startsWith(AppRoutes.notifications)) return 2;
    if (location.startsWith(AppRoutes.settings)) return 3;
    return 0;
  }

  @override
  Widget build(BuildContext context) {
    final selectedIndex = _currentIndex(context);
    final l10n = AppLocalizations.of(context);

    return Scaffold(
      body: child,
      bottomNavigationBar: FBottomNavigationBar(
        index: selectedIndex,
        onChange: (index) {
          switch (index) {
            case 0:
              context.go(AppRoutes.home);
            case 1:
              context.go(AppRoutes.messages);
            case 2:
              context.go(AppRoutes.notifications);
            case 3:
              context.go(AppRoutes.settings);
          }
        },
        children: [
          FBottomNavigationBarItem(
            icon: const Icon(FIcons.house),
            label: Text(l10n.navHome),
          ),
          FBottomNavigationBarItem(
            icon: const Icon(FIcons.messageCircle),
            label: Text(l10n.navMessages),
          ),
          FBottomNavigationBarItem(
            icon: const Icon(FIcons.bell),
            label: Text(l10n.navNotifications),
          ),
          FBottomNavigationBarItem(
            icon: const Icon(FIcons.settings),
            label: Text(l10n.navSettings),
          ),
        ],
      ),
    );
  }
}
