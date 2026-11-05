/// Riverpod providers of notifications (web `/notifications` and the header
/// bell).
///
/// Everything is `autoDispose`, so leaving the signed-in shell on logout drops
/// it and the next account never sees the previous one's data. The shell
/// watches [unreadNotificationCountProvider] (the tab badge) for the whole
/// session; the list lives while its screen does.
///
/// Realtime: [notificationEventsProvider] turns the socket's
/// `notification.created` pushes (and reconnects) into typed events once;
/// the count and the list each listen and apply what concerns them. After a
/// reconnect (also after the app was in the background, where the socket is
/// closed) both load again, and the list merges by id, so nothing shows twice.
library;

import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/features/notifications/data/app_notification.dart';
import 'package:letslive_mobile/features/notifications/data/notification_repository.dart';
import 'package:letslive_mobile/features/notifications/domain/notification_event.dart';

/// Notification REST calls.
final notificationRepositoryProvider = Provider<NotificationRepository>(
  (ref) => NotificationRepository(ref.watch(apiClientProvider)),
);

/// The socket's notification pushes as typed events. Listeners that are not
/// attached miss events, which is intended: a provider that is gone has
/// nothing to update.
final notificationEventsProvider =
    Provider.autoDispose<Stream<NotificationEvent>>((ref) {
      final realtime = ref.read(realtimeManagerProvider);
      final controller = StreamController<NotificationEvent>.broadcast();
      final cleanups = [
        realtime.onEvent(notificationCreatedEvent, (data) {
          try {
            controller.add(NotificationCreated(AppNotification.parse(data)));
          } on FormatException {
            // A payload this app does not understand: ask the server instead.
            controller.add(const NotificationsStale());
          }
        }),
        realtime.onReconnect(() => controller.add(const NotificationsStale())),
      ];
      ref.onDispose(() {
        for (final cleanup in cleanups) {
          cleanup();
        }
        unawaited(controller.close());
      });
      return controller.stream;
    });

/// Listens to [notificationEventsProvider] until the owner of [ref] is
/// disposed.
void _listenToEvents(Ref ref, void Function(NotificationEvent event) onEvent) {
  final subscription = ref.watch(notificationEventsProvider).listen(onEvent);
  ref.onDispose(subscription.cancel);
}

// --- list ------------------------------------------------------------------

/// The user's notifications, newest first, paged.
final notificationsProvider =
    AsyncNotifierProvider.autoDispose<
      NotificationsNotifier,
      PagedFeed<AppNotification>
    >(NotificationsNotifier.new, retry: noRetry);

/// Loads the list, keeps it current from pushes and applies the user's
/// actions after the server accepted them (web refetches after each action;
/// the same result without the extra request).
class NotificationsNotifier extends PagedFeedNotifier<AppNotification> {
  @override
  int get pageSize => notificationsPageSize;

  @override
  String idOf(AppNotification item) => item.id;

  @override
  Future<ApiResponse<List<AppNotification>>> fetchPage(int page) =>
      ref.read(notificationRepositoryProvider).list(page: page);

  @override
  Future<PagedFeed<AppNotification>> build() {
    _listenToEvents(ref, _onEvent);
    return super.build();
  }

  void _onEvent(NotificationEvent event) {
    switch (event) {
      case NotificationsStale():
        unawaited(refreshFirstPage());
      case NotificationCreated(:final notification):
        final feed = state.value;
        if (feed == null || feed.items.any((n) => n.id == notification.id)) {
          return;
        }
        final total = feed.total;
        state = AsyncData(
          feed.copyWith(
            items: [notification, ...feed.items],
            total: total == null ? null : total + 1,
          ),
        );
    }
  }

  /// Marks [id] as read. Returns the failed response, or null on success.
  Future<ApiResponse<Object?>?> markRead(String id) async {
    final feed = state.value;
    final index = feed?.items.indexWhere((n) => n.id == id) ?? -1;
    if (feed == null || index < 0 || feed.items[index].isRead) return null;
    final response = await ref
        .read(notificationRepositoryProvider)
        .markRead(id);
    if (!response.success) return response;
    _replaceItems(
      (items) => [for (final n in items) n.id == id ? n.asRead() : n],
    );
    ref.read(unreadNotificationCountProvider.notifier).decrement();
    return null;
  }

  /// Marks every notification as read. Returns the failed response, or null
  /// on success.
  Future<ApiResponse<Object?>?> markAllRead() async {
    final response = await ref
        .read(notificationRepositoryProvider)
        .markAllRead();
    if (!response.success) return response;
    _replaceItems((items) => [for (final n in items) n.asRead()]);
    ref.read(unreadNotificationCountProvider.notifier).clear();
    return null;
  }

  /// Deletes [id]. Returns the failed response, or null on success.
  Future<ApiResponse<Object?>?> delete(String id) async {
    final feed = state.value;
    final index = feed?.items.indexWhere((n) => n.id == id) ?? -1;
    if (feed == null || index < 0) return null;
    final wasUnread = !feed.items[index].isRead;
    final response = await ref.read(notificationRepositoryProvider).delete(id);
    if (!response.success) return response;
    final current = state.value;
    if (current != null) {
      final total = current.total;
      state = AsyncData(
        current.copyWith(
          items: [
            for (final n in current.items)
              if (n.id != id) n,
          ],
          total: total == null ? null : total - 1,
        ),
      );
    }
    if (wasUnread) {
      ref.read(unreadNotificationCountProvider.notifier).decrement();
    }
    return null;
  }

  void _replaceItems(
    List<AppNotification> Function(List<AppNotification> items) change,
  ) {
    final feed = state.value;
    if (feed == null) return;
    state = AsyncData(feed.copyWith(items: change(feed.items)));
  }
}

// --- unread count ------------------------------------------------------------

/// How many notifications are unread.
final unreadNotificationCountProvider =
    AsyncNotifierProvider.autoDispose<UnreadNotificationsNotifier, int>(
      UnreadNotificationsNotifier.new,
      retry: noRetry,
    );

/// Loads the unread count and counts new notifications from pushes. A count
/// the server sent while the event was in flight is the fresher one: events
/// that arrive before the first load are dropped. A notification counts once
/// however often its event arrives.
class UnreadNotificationsNotifier extends AsyncNotifier<int> {
  /// Ids of the notifications counted from events since the last load.
  final Set<String> _counted = {};

  @override
  Future<int> build() async {
    _counted.clear();
    _listenToEvents(ref, _onEvent);
    return unwrapData(
      await ref.read(notificationRepositoryProvider).unreadCount(),
    );
  }

  /// One notification was read or deleted.
  void decrement() {
    final count = state.value;
    if (count != null && count > 0) state = AsyncData(count - 1);
  }

  /// Everything was read.
  void clear() {
    if (state.hasValue) state = const AsyncData(0);
  }

  void _onEvent(NotificationEvent event) {
    switch (event) {
      case NotificationsStale():
        ref.invalidateSelf();
      case NotificationCreated(:final notification):
        final count = state.value;
        if (count != null &&
            !notification.isRead &&
            _counted.add(notification.id)) {
          state = AsyncData(count + 1);
        }
    }
  }
}

/// The Notifications tab badge. The shell watches it for the whole session,
/// which keeps the count (and so its realtime listener) alive.
final notificationBadgeProvider = Provider.autoDispose<int>(
  (ref) => ref.watch(unreadNotificationCountProvider).value ?? 0,
);
