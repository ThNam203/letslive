import 'package:flutter/material.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/features/home/presentation/home_screen.dart';
import 'package:letslive_mobile/features/search/presentation/search_screen.dart';

/// One tab of the bottom navigation: where it lives, how it looks, what it
/// shows.
class AppTab {
  const AppTab({
    required this.path,
    required this.icon,
    required this.label,
    required this.builder,
  });

  /// Route path of the tab's root screen, e.g. `/search`.
  final String path;

  /// Icon of the navigation bar destination.
  final IconData icon;

  /// Localized destination label.
  final String Function(AppLocalizations l10n) label;

  /// Builds the tab's root screen.
  final Widget Function(BuildContext context) builder;
}

/// The tab registry, in display order. The final order is Home, Search,
/// Notifications, Messages, Profile.
///
/// To add a tab, a task inserts one [AppTab] at its place in this list
/// (nothing else to wire: [AppShell] builds the navigation bar and the
/// router builds the branches from it). No placeholder tabs: a tab appears
/// only when its screen exists. The bar is hidden while fewer than two tabs
/// are registered.
final List<AppTab> appTabs = [
  AppTab(
    path: AppRoutes.home,
    icon: Icons.home_outlined,
    label: (l10n) => l10n.common_home,
    builder: (context) => const HomeScreen(),
  ),
  AppTab(
    path: AppRoutes.search,
    icon: Icons.search,
    label: (l10n) => l10n.common_search_title,
    builder: (context) => const SearchScreen(),
  ),
];
