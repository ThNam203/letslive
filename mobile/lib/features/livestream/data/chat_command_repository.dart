import '../../../core/network/api_client.dart';
import '../../../core/network/api_endpoints.dart';
import '../../../core/network/api_response.dart';
import '../../../models/chat_command.dart';

class ChatCommandRepository {
  final ApiClient _client;

  ChatCommandRepository(this._client);

  /// Returns the merged set of chat commands available in the given room:
  /// channel-scope commands for the room owner plus the caller's user-scope
  /// commands when authenticated.
  Future<ApiResponse<List<ChatCommand>>> listForRoom(String roomId) {
    return _client.get(
      ApiEndpoints.chatCommands,
      queryParameters: {'roomId': roomId},
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => ChatCommand.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<MyChatCommands>> listMine() {
    return _client.get(
      ApiEndpoints.chatCommandsMine,
      fromJsonT: (json) => MyChatCommands.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<ChatCommand>> create({
    required ChatCommandScope scope,
    required String name,
    required String response,
    String description = '',
  }) {
    return _client.post(
      ApiEndpoints.chatCommands,
      data: {
        'scope': chatCommandScopeToString(scope),
        'name': name,
        'response': response,
        'description': description,
      },
      fromJsonT: (json) => ChatCommand.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<ChatCommand>> update({
    required String id,
    String? name,
    String? response,
    String? description,
  }) {
    final body = <String, dynamic>{};
    if (name != null) body['name'] = name;
    if (response != null) body['response'] = response;
    if (description != null) body['description'] = description;
    return _client.patch(
      ApiEndpoints.chatCommandById(id),
      data: body,
      fromJsonT: (json) => ChatCommand.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<void>> delete(String id) {
    return _client.delete(ApiEndpoints.chatCommandById(id));
  }
}
