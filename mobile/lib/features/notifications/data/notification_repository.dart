import '../../../core/network/api_client.dart';
import '../../../core/network/api_endpoints.dart';
import '../../../core/network/api_response.dart';
import '../../../models/notification.dart';

class NotificationRepository {
  final ApiClient _client;

  NotificationRepository(this._client);

  Future<ApiResponse<List<AppNotification>>> getNotifications({
    int page = 0,
    int pageSize = 20,
  }) {
    return _client.get(
      ApiEndpoints.notifications,
      queryParameters: {'page': page, 'page_size': pageSize},
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => AppNotification.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<int>> getUnreadCount() {
    return _client.get(
      ApiEndpoints.notificationsUnreadCount,
      fromJsonT: (json) => json as int,
    );
  }

  Future<ApiResponse<void>> markAsRead(String id) {
    return _client.patch(ApiEndpoints.notificationRead(id));
  }

  Future<ApiResponse<void>> markAllAsRead() {
    return _client.patch(ApiEndpoints.notificationsReadAll);
  }

  Future<ApiResponse<void>> deleteNotification(String id) {
    return _client.delete(ApiEndpoints.notificationById(id));
  }
}
