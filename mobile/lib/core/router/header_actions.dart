import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/features/notifications/notifications_providers.dart';

/// Search and the notification bell, at the right end of every tab's app bar
/// (web keeps both in its top navbar). Each opens its screen over the tabs.
///
/// Put it last in `AppBar.actions`, so the bell sits in the top-right corner
/// on every tab.
class HeaderActions extends ConsumerWidget {
  const HeaderActions({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final unread = ref.watch(notificationBadgeProvider);
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        IconButton(
          key: const ValueKey('open-search'),
          tooltip: l10n.common_search_title,
          icon: const Icon(Icons.search),
          onPressed: () => context.push(AppRoutes.search),
        ),
        IconButton(
          key: const ValueKey('open-notifications'),
          tooltip: l10n.notification_title,
          icon: Badge.count(
            count: unread,
            isLabelVisible: unread > 0,
            child: const Icon(Icons.notifications_none),
          ),
          onPressed: () => context.push(AppRoutes.notifications),
        ),
      ],
    );
  }
}
