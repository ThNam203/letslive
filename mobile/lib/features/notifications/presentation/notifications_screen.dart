/// The notifications screen (web `/notifications`), opened from the header
/// bell: the user's notifications,
/// newest first, unread ones marked. Tapping a row marks it read and opens
/// what it is about; a menu on the row marks it read or deletes it, and the
/// app bar marks everything read.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/push/push_tap.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/core/widgets/action_feedback.dart';
import 'package:letslive_mobile/core/widgets/feed_section.dart';
import 'package:letslive_mobile/features/notifications/data/app_notification.dart';
import 'package:letslive_mobile/features/notifications/domain/time_ago.dart';
import 'package:letslive_mobile/features/notifications/notifications_providers.dart';

/// The notifications screen.
class NotificationsScreen extends ConsumerWidget {
  const NotificationsScreen({super.key});

  /// The retry of a failed load: the list and the count.
  static void _reload(WidgetRef ref) {
    ref
      ..invalidate(unreadNotificationCountProvider)
      ..invalidate(notificationsProvider);
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final notifications = ref.watch(notificationsProvider);
    final feed = notifications.value;
    final hasUnread =
        ref.watch(notificationBadgeProvider) > 0 ||
        (feed?.items.any((n) => !n.isRead) ?? false);
    final status = feedStatusBody(
      value: notifications.whenData((feed) => feed.items),
      onRetry: () => _reload(ref),
      empty: EmptyNotice(
        icon: Icons.notifications_none,
        title: l10n.notification_no_notifications_yet,
      ),
    );
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.notification_title),
        actions: [
          if (hasUnread)
            // An icon, not a text button: a label would overflow the bar on a
            // narrow screen with large text.
            IconButton(
              key: const ValueKey('mark-all-read'),
              tooltip: l10n.notification_mark_all_as_read,
              icon: const Icon(Icons.done_all),
              onPressed: () async {
                final failure = await ref
                    .read(notificationsProvider.notifier)
                    .markAllRead();
                if (context.mounted) showActionError(context, failure);
              },
            ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () {
          // The count too: a failed load of it would otherwise stay failed
          // for the whole session.
          ref.invalidate(unreadNotificationCountProvider);
          return ref.refresh(notificationsProvider.future);
        },
        child: status != null
            ? ListView(children: [status])
            : ListView.builder(
                itemCount: feed!.items.length + 1,
                itemBuilder: (context, index) => index == feed.items.length
                    ? PagedFooter<AppNotification>(
                        feed: feed,
                        onLoadMore: () =>
                            ref.read(notificationsProvider.notifier).loadMore(),
                      )
                    : NotificationTile(
                        key: ValueKey('notification-${feed.items[index].id}'),
                        notification: feed.items[index],
                      ),
              ),
      ),
    );
  }
}

/// What a row's menu offers.
enum _RowAction { markRead, delete }

/// One row. Text is clipped with an ellipsis so a long title or message
/// never overflows, even on a narrow screen with large text.
class NotificationTile extends ConsumerWidget {
  const NotificationTile({required this.notification, super.key});

  final AppNotification notification;

  Future<void> _open(BuildContext context, WidgetRef ref) async {
    final route = resolveDeepLink(
      PushTap(
        type: notification.type,
        referenceId: notification.referenceId,
        actionUrl: notification.actionUrl,
      ),
    );
    final marked = ref
        .read(notificationsProvider.notifier)
        .markRead(notification.id);
    if (route != AppRoutes.notifications) await context.push<void>(route);
    final failure = await marked;
    if (context.mounted) showActionError(context, failure);
  }

  Future<void> _onAction(
    BuildContext context,
    WidgetRef ref,
    _RowAction action,
  ) async {
    final notifier = ref.read(notificationsProvider.notifier);
    final ApiResponse<Object?>? failure = switch (action) {
      _RowAction.markRead => await notifier.markRead(notification.id),
      _RowAction.delete => await notifier.delete(notification.id),
    };
    if (context.mounted) showActionError(context, failure);
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final muted = theme.colorScheme.onSurfaceVariant;
    final label = notification.actionLabel;
    return InkWell(
      onTap: () => _open(context, ref),
      child: Container(
        color: notification.isRead
            ? null
            : theme.colorScheme.primary.withValues(alpha: 0.05),
        padding: const EdgeInsets.only(left: 16, top: 12, bottom: 12),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SizedBox(
              width: 8,
              height: 24,
              child: notification.isRead
                  ? null
                  : Center(
                      child: DecoratedBox(
                        decoration: BoxDecoration(
                          color: theme.colorScheme.primary,
                          shape: BoxShape.circle,
                        ),
                        child: const SizedBox.square(dimension: 8),
                      ),
                    ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    notification.title,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: theme.textTheme.titleSmall,
                  ),
                  Text(
                    notification.message,
                    maxLines: 3,
                    overflow: TextOverflow.ellipsis,
                    style: theme.textTheme.bodyMedium?.copyWith(color: muted),
                  ),
                  Text(
                    notificationTimeAgo(
                      notification.createdAt,
                      DateTime.now(),
                      l10n,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: theme.textTheme.bodySmall?.copyWith(color: muted),
                  ),
                  if (label != null && label.isNotEmpty)
                    Text(
                      label,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: theme.textTheme.labelMedium?.copyWith(
                        color: theme.colorScheme.primary,
                      ),
                    ),
                ],
              ),
            ),
            PopupMenuButton<_RowAction>(
              key: ValueKey('notification-menu-${notification.id}'),
              onSelected: (action) => _onAction(context, ref, action),
              itemBuilder: (context) => [
                if (!notification.isRead)
                  PopupMenuItem(
                    value: _RowAction.markRead,
                    child: Text(l10n.notification_mark_as_read),
                  ),
                PopupMenuItem(
                  value: _RowAction.delete,
                  child: Text(l10n.notification_delete),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
