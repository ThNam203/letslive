/// Realtime notification events, after `use-notification-realtime.ts`.
/// The server delivers `notification.created` on the user's own topic, which
/// it joins for the socket, so (like DMs, and like web) nothing is subscribed.
library;

import 'package:letslive_mobile/features/notifications/data/app_notification.dart';

/// The event type web listens to (`REALTIME_EVENT.NOTIFICATION_CREATED`).
const String notificationCreatedEvent = 'notification.created';

/// Something that changed the user's notifications.
sealed class NotificationEvent {
  const NotificationEvent();
}

/// A notification was created for the user.
class NotificationCreated extends NotificationEvent {
  const NotificationCreated(this.notification);

  final AppNotification notification;
}

/// Events may have been missed (the socket reconnected, or a payload was not
/// understood): load the list and the count again.
class NotificationsStale extends NotificationEvent {
  const NotificationsStale();
}
