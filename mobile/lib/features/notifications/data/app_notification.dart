import 'package:flutter/foundation.dart';
import 'package:letslive_mobile/core/network/json_fields.dart';

/// One notification of the user, typed after `Notification` in
/// `web/types/notification.ts` (named `AppNotification` because Flutter has a
/// `Notification` class). Also the payload of the realtime event
/// `notification.created`.
@immutable
class AppNotification {
  const AppNotification({
    required this.id,
    required this.type,
    required this.title,
    required this.message,
    required this.isRead,
    required this.createdAt,
    this.actionUrl,
    this.actionLabel,
    this.referenceId,
  });

  factory AppNotification.fromJson(Map<String, Object?> json) {
    final f = JsonFields('AppNotification', json);
    return AppNotification(
      id: f.string('id'),
      type: f.string('type'),
      title: f.string('title'),
      message: f.string('message'),
      isRead: f.boolean('isRead'),
      createdAt: f.date('createdAt'),
      actionUrl: f.optString('actionUrl'),
      actionLabel: f.optString('actionLabel'),
      referenceId: f.optString('referenceId'),
    );
  }

  /// `ParseData<AppNotification>` for `ApiClient` and the realtime payload.
  static AppNotification parse(Object? data) =>
      parseJsonObject(data, 'AppNotification', AppNotification.fromJson);

  /// `ParseData<List<AppNotification>>` for `ApiClient`.
  static List<AppNotification> parseList(Object? data) =>
      parseJsonList(data, 'AppNotification', AppNotification.fromJson);

  /// The notification's id.
  final String id;

  /// What happened: `dm`, `vod_comment`, `vod_comment_reply`,
  /// `gift_received`, ...
  final String type;

  /// Short headline, as stored (not localized).
  final String title;

  /// The text under the title, as stored (not localized).
  final String message;

  /// Whether the user has read it.
  final bool isRead;

  /// When it was created.
  final DateTime createdAt;

  /// A web path to open, e.g. `/users/<id>/vods/<id>`; see `resolveDeepLink`.
  final String? actionUrl;

  /// Text of the link to [actionUrl] ("View" when null).
  final String? actionLabel;

  /// Id of the thing it is about.
  final String? referenceId;

  /// This notification marked as read.
  AppNotification asRead() => AppNotification(
    id: id,
    type: type,
    title: title,
    message: message,
    isRead: true,
    createdAt: createdAt,
    actionUrl: actionUrl,
    actionLabel: actionLabel,
    referenceId: referenceId,
  );
}
