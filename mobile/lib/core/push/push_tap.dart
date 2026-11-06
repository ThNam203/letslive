/// What a tapped push notification says about where to go, and the one
/// function that turns it into a screen of the app.
library;

import 'package:flutter/foundation.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';

/// The deep-link data the backend attaches to a push (`PushDataDTO` in
/// `backend/user/dto/push.go`; stored notifications carry the same three
/// fields). Everything is optional on the wire; a push with none of it
/// still opens the app, at the notifications screen.
@immutable
class PushTap {
  const PushTap({this.type = '', this.referenceId, this.actionUrl});

  /// Reads the data map of an FCM message. Values that are not strings are
  /// ignored, so a malformed push can only land on the fallback screen.
  factory PushTap.fromData(Map<String, Object?> data) {
    String? text(String key) {
      final value = data[key];
      return value is String && value.isNotEmpty ? value : null;
    }

    return PushTap(
      type: text('type') ?? '',
      referenceId: text('referenceId'),
      actionUrl: text('actionUrl'),
    );
  }

  /// What happened: `dm`, `vod_comment`, `vod_comment_reply`,
  /// `gift_received`, ...; empty when unknown.
  final String type;

  /// Id of the thing it is about (the conversation for `dm`, a comment or a
  /// gift for the others).
  final String? referenceId;

  /// A web path such as `/messages/<id>` or `/users/<id>/vods/<id>`.
  final String? actionUrl;
}

/// The app route [tap] opens.
///
/// The `actionUrl` wins when it is one of the web paths the app has a
/// screen for. Otherwise the `type` decides (a DM names its conversation in
/// `referenceId`). Anything else, an unknown or malformed url included,
/// falls back to [AppRoutes.notifications]: a push (or a notification row)
/// can never send the app outside its own routes, because every route is
/// rebuilt here from path parts, never copied from the url.
String resolveDeepLink(PushTap tap) {
  final fromUrl = _routeOfActionUrl(tap.actionUrl);
  if (fromUrl != null) return fromUrl;
  final reference = tap.referenceId;
  if (tap.type == 'dm' && reference != null) {
    return AppRoutes.conversation(reference);
  }
  return AppRoutes.notifications;
}

/// The route of a web path, or null when it is not a path of the app.
String? _routeOfActionUrl(String? actionUrl) {
  if (actionUrl == null || !actionUrl.startsWith('/')) return null;
  final uri = Uri.tryParse(actionUrl);
  // `//host/path` is a network-path reference, not an app path.
  if (uri == null || uri.hasAuthority || actionUrl.startsWith('//')) {
    return null;
  }
  final parts = uri.pathSegments;
  // Empty and dot segments are never ids; rejecting them also keeps `..`
  // from walking out of the path.
  if (parts.any((part) => part.isEmpty || part == '.' || part == '..')) {
    return null;
  }
  return switch (parts) {
    ['messages', final id] when id != 'new' => AppRoutes.conversation(id),
    ['users', final id] => AppRoutes.channel(id),
    ['users', final userId, 'vods', final vodId] => AppRoutes.vod(
      userId,
      vodId,
    ),
    _ => null,
  };
}
