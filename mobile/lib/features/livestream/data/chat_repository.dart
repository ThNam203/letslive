import '../../../core/network/api_client.dart';
import '../../../core/network/api_endpoints.dart';
import '../../../core/network/api_response.dart';
import '../../../models/chat_message.dart';

class ChatRepository {
  final ApiClient _client;

  ChatRepository(this._client);

  Future<ApiResponse<List<ChatMessage>>> getMessages({required String roomId}) {
    return _client.get(
      ApiEndpoints.chatMessages,
      queryParameters: {'roomId': roomId},
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => ChatMessage.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}
