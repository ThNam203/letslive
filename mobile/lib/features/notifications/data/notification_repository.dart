import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/network/json_fields.dart';
import 'package:letslive_mobile/features/notifications/data/app_notification.dart';

/// Notifications per page; the endpoint takes no limit (the backend pages 20
/// at a time, `web/hooks/queries/use-notifications.ts`).
const int notificationsPageSize = 20;

/// The `/user/me/notifications` endpoints, mirroring
/// `web/lib/api/notification.ts`. Failures are values
/// ([ApiResponse.success] false), never thrown.
class NotificationRepository {
  const NotificationRepository(this._client);

  final ApiClient _client;

  static const String _base = ApiEndpoints.notifications;

  /// `GET /user/me/notifications`: [notificationsPageSize] per [page]
  /// (0-based), newest first.
  Future<ApiResponse<List<AppNotification>>> list({int page = 0}) => _client
      .get(_base, query: {'page': page}, parse: AppNotification.parseList);

  /// `GET /user/me/notifications/unread-count`.
  Future<ApiResponse<int>> unreadCount() => _client.get(
    '$_base/unread-count',
    parse: (data) => parseJsonObject(
      data,
      'UnreadCount',
      (json) => JsonFields('UnreadCount', json).integer('count'),
    ),
  );

  /// `PATCH /user/me/notifications/<id>/read`.
  Future<ApiResponse<Object?>> markRead(String id) => _client.patch(
    '$_base/${Uri.encodeComponent(id)}/read',
    parse: (data) => data,
  );

  /// `PATCH /user/me/notifications/read-all`.
  Future<ApiResponse<Object?>> markAllRead() =>
      _client.patch('$_base/read-all', parse: (data) => data);

  /// `DELETE /user/me/notifications/<id>`.
  Future<ApiResponse<Object?>> delete(String id) => _client.delete(
    '$_base/${Uri.encodeComponent(id)}',
    parse: (data) => data,
  );
}
