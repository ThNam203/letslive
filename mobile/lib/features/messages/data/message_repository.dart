import '../../../core/network/api_client.dart';
import '../../../core/network/api_endpoints.dart';
import '../../../core/network/api_response.dart';
import '../../../models/conversation.dart';

class MessageRepository {
  final ApiClient _client;

  MessageRepository(this._client);

  Future<ApiResponse<List<Conversation>>> getConversations({
    int page = 0,
    int limit = 20,
  }) {
    return _client.get(
      ApiEndpoints.conversations,
      queryParameters: {
        'page': page,
        'limit': limit,
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
    String? before,
    int limit = 50,
  }) {
    return _client.get(
      ApiEndpoints.conversationMessages(id),
      queryParameters: {
        'limit': limit,
        'before': ?before,
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
    return _client.post(ApiEndpoints.conversationRead(id));
  }

  Future<ApiResponse<DmMessage>> sendMessage(
    String conversationId, {
    required String text,
    required String senderUsername,
    String type = 'text',
    List<String>? imageUrls,
    String? replyTo,
  }) {
    return _client.post(
      ApiEndpoints.conversationMessages(conversationId),
      data: {
        'text': text,
        'type': type,
        'senderUsername': senderUsername,
        'imageUrls': ?imageUrls,
        'replyTo': ?replyTo,
      },
      fromJsonT: (json) =>
          DmMessage.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<DmMessage>> editMessage(
    String conversationId,
    String messageId, {
    required String text,
  }) {
    return _client.patch(
      '${ApiEndpoints.conversationMessages(conversationId)}/$messageId',
      data: {'text': text},
      fromJsonT: (json) =>
          DmMessage.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<void>> deleteMessage(
    String conversationId,
    String messageId,
  ) {
    return _client.delete(
      '${ApiEndpoints.conversationMessages(conversationId)}/$messageId',
    );
  }

  Future<ApiResponse<Conversation>> updateConversation(
    String id, {
    String? name,
    String? avatarUrl,
  }) {
    return _client.put(
      ApiEndpoints.conversationById(id),
      data: {
        'name': ?name,
        'avatarUrl': ?avatarUrl,
      },
      fromJsonT: (json) =>
          Conversation.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<void>> leaveConversation(String id) {
    return _client.delete(ApiEndpoints.conversationById(id));
  }

  Future<ApiResponse<void>> addParticipant(
    String conversationId, {
    required String userId,
    required String username,
    String? displayName,
    String? profilePicture,
  }) {
    return _client.post(
      ApiEndpoints.conversationParticipants(conversationId),
      data: {
        'userId': userId,
        'username': username,
        'displayName': ?displayName,
        'profilePicture': ?profilePicture,
      },
    );
  }

  Future<ApiResponse<void>> removeParticipant(
    String conversationId,
    String userId,
  ) {
    return _client.delete(
      ApiEndpoints.conversationRemoveParticipant(conversationId, userId),
    );
  }
}
