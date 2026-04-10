import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../l10n/app_localizations.dart';
import '../../../models/notification.dart';
import '../../../providers.dart';
import '../../../shared/widgets/empty_state_view.dart';
import '../../../shared/widgets/error_display.dart';
import '../../../shared/widgets/loading_indicator.dart';

class NotificationsScreen extends ConsumerStatefulWidget {
  const NotificationsScreen({super.key});

  @override
  ConsumerState<NotificationsScreen> createState() =>
      _NotificationsScreenState();
}

class _NotificationsScreenState extends ConsumerState<NotificationsScreen> {
  List<AppNotification> _notifications = [];
  bool _isLoading = true;
  String? _error;
  bool _isMarkingAllRead = false;
  int _currentPage = 0;
  bool _hasMore = true;
  bool _isLoadingMore = false;

  @override
  void initState() {
    super.initState();
    _fetchNotifications();
  }

  Future<void> _fetchNotifications() async {
    setState(() {
      _isLoading = true;
      _error = null;
      _currentPage = 0;
    });

    try {
      final repo = ref.read(notificationRepositoryProvider);
      final response = await repo.getNotifications(page: 0);
      if (!mounted) return;

      if (response.success) {
        final pageSize = response.meta?.pageSize ?? 20;
        final total = response.meta?.total ?? 0;
        final totalPages = (total + pageSize - 1) ~/ pageSize;
        setState(() {
          _notifications = response.data ?? [];
          _isLoading = false;
          _hasMore = (_currentPage + 1) < totalPages;
        });
        ref.read(unreadNotificationCountProvider.notifier).fetch();
      } else {
        setState(() {
          _error = response.message;
          _isLoading = false;
        });
      }
    } on DioException catch (_) {
      if (mounted) {
        setState(() {
          _error = AppLocalizations.of(context).fetchError;
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _loadMore() async {
    if (_isLoadingMore || !_hasMore) return;

    setState(() => _isLoadingMore = true);

    try {
      final repo = ref.read(notificationRepositoryProvider);
      final nextPage = _currentPage + 1;
      final response = await repo.getNotifications(page: nextPage);
      if (!mounted) return;

      if (response.success) {
        final pageSize = response.meta?.pageSize ?? 20;
        final total = response.meta?.total ?? 0;
        final totalPages = (total + pageSize - 1) ~/ pageSize;
        setState(() {
          _currentPage = nextPage;
          _notifications.addAll(response.data ?? []);
          _isLoadingMore = false;
          _hasMore = (_currentPage + 1) < totalPages;
        });
      } else {
        setState(() => _isLoadingMore = false);
      }
    } on DioException catch (_) {
      if (mounted) setState(() => _isLoadingMore = false);
    }
  }

  Future<void> _markAllAsRead() async {
    setState(() => _isMarkingAllRead = true);

    try {
      final repo = ref.read(notificationRepositoryProvider);
      final response = await repo.markAllAsRead();

      if (!mounted) return;

      if (response.success) {
        setState(() {
          _notifications = _notifications
              .map((n) => n.copyWith(isRead: true))
              .toList();
        });
        ref.read(unreadNotificationCountProvider.notifier).clear();
      }
    } on DioException catch (_) {
      if (mounted) {
        final l10n = AppLocalizations.of(context);
        showFToast(
          context: context,
          title: Text(l10n.fetchError),
          icon: const Icon(FIcons.circleAlert),
        );
      }
    } finally {
      if (mounted) setState(() => _isMarkingAllRead = false);
    }
  }

  Future<void> _markAsRead(AppNotification notification) async {
    if (notification.isRead) return;

    try {
      final repo = ref.read(notificationRepositoryProvider);
      final response = await repo.markAsRead(notification.id);

      if (!mounted) return;

      if (response.success) {
        setState(() {
          final index = _notifications.indexWhere(
            (n) => n.id == notification.id,
          );
          if (index != -1) {
            _notifications[index] = notification.copyWith(isRead: true);
          }
        });
        ref.read(unreadNotificationCountProvider.notifier).decrement();
      }
    } on DioException catch (_) {
      // Silent fail for mark as read
    }
  }

  Future<void> _deleteNotification(AppNotification notification) async {
    try {
      final repo = ref.read(notificationRepositoryProvider);
      final response = await repo.deleteNotification(notification.id);

      if (!mounted) return;

      if (response.success) {
        setState(() {
          _notifications.removeWhere((n) => n.id == notification.id);
        });
        if (!notification.isRead) {
          ref.read(unreadNotificationCountProvider.notifier).decrement();
        }
      } else {
        final l10n = AppLocalizations.of(context);
        showFToast(
          context: context,
          title: Text(l10n.fetchError),
          icon: const Icon(FIcons.circleAlert),
        );
      }
    } on DioException catch (_) {
      if (mounted) {
        final l10n = AppLocalizations.of(context);
        showFToast(
          context: context,
          title: Text(l10n.fetchError),
          icon: const Icon(FIcons.circleAlert),
        );
      }
    }
  }

  String _formatTimeAgo(String createdAt, AppLocalizations l10n) {
    final date = DateTime.tryParse(createdAt);
    if (date == null) return '';

    final now = DateTime.now();
    final diff = now.difference(date);

    if (diff.inSeconds < 60) return l10n.timeSecondsAgo(diff.inSeconds);
    if (diff.inMinutes < 60) return l10n.timeMinutesAgo(diff.inMinutes);
    if (diff.inHours < 24) return l10n.timeHoursAgo(diff.inHours);
    if (diff.inDays < 7) return l10n.timeDaysAgo(diff.inDays);
    if (diff.inDays < 30) return l10n.timeWeeksAgo(diff.inDays ~/ 7);
    if (diff.inDays < 365) return l10n.timeMonthsAgo(diff.inDays ~/ 30);
    return l10n.timeYearsAgo(diff.inDays ~/ 365);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final hasUnread = _notifications.any((n) => !n.isRead);

    return FScaffold(
      header: FHeader(
        title: Text(l10n.notificationsTitle),
        suffixes: [
          if (!_isLoading && _notifications.isNotEmpty && hasUnread)
            FButton(
              variant: FButtonVariant.ghost,
              onPress: _isMarkingAllRead ? null : _markAllAsRead,
              child: _isMarkingAllRead
                  ? const SizedBox(
                      height: 16,
                      width: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Text(l10n.notificationsMarkAllAsRead),
            ),
        ],
      ),
      child: _buildContent(l10n),
    );
  }

  Widget _buildContent(AppLocalizations l10n) {
    if (_isLoading) {
      return LoadingIndicator(message: l10n.notificationsLoading);
    }

    if (_error != null) {
      return ErrorDisplay(
        title: l10n.errorGeneralTitle,
        message: _error,
        onRetry: _fetchNotifications,
      );
    }

    if (_notifications.isEmpty) {
      return EmptyStateView(
        icon: FIcons.bell,
        title: l10n.notificationsNoNotifications,
        description: l10n.notificationsNoNotificationsYet,
      );
    }

    return RefreshIndicator(
      onRefresh: _fetchNotifications,
      child: ListView.builder(
        padding: const EdgeInsets.symmetric(vertical: 8),
        itemCount: _notifications.length + (_hasMore ? 1 : 0),
        itemBuilder: (context, index) {
          if (index == _notifications.length) {
            return Padding(
              padding: const EdgeInsets.all(16),
              child: FButton(
                variant: FButtonVariant.outline,
                onPress: _isLoadingMore ? null : _loadMore,
                child: _isLoadingMore
                    ? const SizedBox(
                        height: 16,
                        width: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : Text(l10n.notificationsLoadMore),
              ),
            );
          }

          final notification = _notifications[index];
          return _NotificationTile(
            notification: notification,
            timeAgo: _formatTimeAgo(notification.createdAt, l10n),
            onTap: () => _markAsRead(notification),
            onMarkAsRead: () => _markAsRead(notification),
            onDelete: () => _deleteNotification(notification),
          );
        },
      ),
    );
  }
}

class _NotificationTile extends StatelessWidget {
  final AppNotification notification;
  final String timeAgo;
  final VoidCallback onTap;
  final VoidCallback onMarkAsRead;
  final VoidCallback onDelete;

  const _NotificationTile({
    required this.notification,
    required this.timeAgo,
    required this.onTap,
    required this.onMarkAsRead,
    required this.onDelete,
  });

  IconData _iconForType(String type) {
    switch (type) {
      case 'follow':
        return FIcons.userPlus;
      case 'livestream':
        return FIcons.video;
      case 'vod':
        return FIcons.film;
      case 'comment':
        return FIcons.messageCircle;
      case 'like':
        return FIcons.heart;
      default:
        return FIcons.bell;
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: notification.isRead
              ? null
              : colors.primary.withValues(alpha: 0.05),
          border: Border(bottom: BorderSide(color: colors.border, width: 0.5)),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Icon
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: notification.isRead
                    ? colors.muted
                    : colors.primary.withValues(alpha: 0.1),
                shape: BoxShape.circle,
              ),
              child: Icon(
                _iconForType(notification.type),
                size: 18,
                color: notification.isRead
                    ? colors.mutedForeground
                    : colors.primary,
              ),
            ),
            const SizedBox(width: 12),
            // Content
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    notification.title,
                    style: typography.sm.copyWith(
                      fontWeight: notification.isRead
                          ? FontWeight.normal
                          : FontWeight.w600,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 2),
                  Text(
                    notification.message,
                    style: typography.xs.copyWith(
                      color: colors.mutedForeground,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 4),
                  Text(
                    timeAgo,
                    style: typography.xs.copyWith(
                      color: colors.mutedForeground,
                    ),
                  ),
                ],
              ),
            ),
            // Actions
            PopupMenuButton<String>(
              icon: Icon(
                FIcons.ellipsis,
                size: 18,
                color: colors.mutedForeground,
              ),
              onSelected: (value) {
                if (value == 'read') onMarkAsRead();
                if (value == 'delete') onDelete();
              },
              itemBuilder: (context) => [
                if (!notification.isRead)
                  PopupMenuItem(
                    value: 'read',
                    child: Row(
                      children: [
                        const Icon(FIcons.check, size: 16),
                        const SizedBox(width: 8),
                        Text(l10n.notificationsMarkAsRead),
                      ],
                    ),
                  ),
                PopupMenuItem(
                  value: 'delete',
                  child: Row(
                    children: [
                      Icon(FIcons.trash, size: 16, color: colors.destructive),
                      const SizedBox(width: 8),
                      Text(
                        l10n.notificationsDelete,
                        style: TextStyle(color: colors.destructive),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
