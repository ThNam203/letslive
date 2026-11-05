import 'package:flutter/material.dart';
import 'package:flutter_riverpod/misc.dart' show ProviderListenable;
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/features/home/presentation/home_screen.dart';
import 'package:letslive_mobile/features/messages/messages_providers.dart';
import 'package:letslive_mobile/features/messages/presentation/conversations_screen.dart';
import 'package:letslive_mobile/features/notifications/notifications_providers.dart';
import 'package:letslive_mobile/features/notifications/presentation/notifications_screen.dart';
import 'package:letslive_mobile/features/profile/presentation/own_profile_screen.dart';
import 'package:letslive_mobile/features/search/presentation/search_screen.dart';

/// One tab of the bottom navigation: where it lives, how it looks, what it
/// shows.
class AppTab {
  const AppTab({
    required this.path,
    required this.icon,
    required this.label,
    required this.builder,
    this.badge,
  });

  /// Route path of the tab's root screen, e.g. `/search`.
  final String path;

  /// Icon of the navigation bar destination.
  final IconData icon;

  /// Localized destination label.
  final String Function(AppLocalizations l10n) label;

  /// Builds the tab's root screen.
  final Widget Function(BuildContext context) builder;

  /// How many things need attention on this tab (e.g. unread messages),
  /// shown as a badge on its icon while above zero. Null for no badge. The
  /// shell watches it, so the provider lives as long as the signed-in app.
  final ProviderListenable<int>? badge;
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
  AppTab(
    path: AppRoutes.notifications,
    icon: Icons.notifications_none,
    label: (l10n) => l10n.notification_title,
    builder: (context) => const NotificationsScreen(),
    badge: notificationBadgeProvider,
  ),
  AppTab(
    path: AppRoutes.messages,
    icon: Icons.chat_bubble_outline,
    label: (l10n) => l10n.messages_title,
    builder: (context) => const ConversationsScreen(),
    badge: unreadDmCountProvider,
  ),
  AppTab(
    path: AppRoutes.profile,
    icon: Icons.person_outline,
    label: (l10n) => l10n.settings_navigation_profile,
    builder: (context) => const OwnProfileScreen(),
  ),
];
