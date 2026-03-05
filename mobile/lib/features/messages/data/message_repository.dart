import '../../../core/network/api_client.dart';
import '../../../core/network/api_endpoints.dart';
import '../../../core/network/api_response.dart';
import '../../../models/conversation.dart';

class MessageRepository {
  final ApiClient _client;

  MessageRepository(this._client);

  Future<ApiResponse<List<Conversation>>> getConversations({
    int page = 0,
    int pageSize = 20,
  }) {
    return _client.get(
      ApiEndpoints.conversations,
      queryParameters: {
        'page': page,
        'page_size': pageSize,
      },
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => Conversation.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<Map<String, dynamic>>> getUnreadCounts() {
    return _client.get(
      ApiEndpoints.conversationsUnreadCounts,
      fromJsonT: (json) => json as Map<String, dynamic>,
    );
  }

  Future<ApiResponse<Conversation>> getConversation(String id) {
    return _client.get(
      ApiEndpoints.conversationById(id),
      fromJsonT: (json) =>
          Conversation.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<List<DmMessage>>> getConversationMessages(
    String id, {
    int page = 0,
    int pageSize = 50,
  }) {
    return _client.get(
      ApiEndpoints.conversationMessages(id),
      queryParameters: {
        'page': page,
        'page_size': pageSize,
      },
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => DmMessage.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<void>> createConversation({
    required List<String> participantIds,
    String? name,
  }) {
    return _client.post(
      ApiEndpoints.conversations,
      data: {
        'participantIds': participantIds,
        // ignore: use_null_aware_elements
        if (name != null) 'name': name,
      },
    );
  }

  Future<ApiResponse<void>> markConversationRead(String id) {
    return _client.patch(ApiEndpoints.conversationRead(id));
  }
}
