import 'package:letslive_mobile/core/chat_parser/chat_command.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/watch/chat/data/chat_message.dart';

/// REST side of the livestream chat (`web/lib/api/chat.ts`,
/// `chat-command.ts`). These are the room's lines, not direct messages. The
/// room id is the streamer's user id. Failures are values
/// ([ApiResponse.success] false), never thrown.
class ChatRepository {
  const ChatRepository(this._client);

  final ApiClient _client;

  /// `GET /messages?roomId=`: the room's latest lines, oldest first.
  Future<ApiResponse<List<ChatMessage>>> history(String roomId) => _client.get(
    ApiEndpoints.messages,
    query: {'roomId': roomId},
    parse: ChatMessage.parseList,
  );

  /// `POST /messages`: says [text] in the room as the signed-in user. The
  /// line also reaches every viewer, the sender included, as a
  /// `chat.message` push carrying the same id as the response.
  Future<ApiResponse<ChatMessage>> send(String roomId, String text) =>
      _client.post(
        ApiEndpoints.messages,
        body: {'roomId': roomId, 'text': text},
        parse: ChatMessage.parse,
      );

  /// `GET /chat-commands?roomId=`: the room's channel commands plus, when
  /// signed in, the viewer's own.
  Future<ApiResponse<List<ChatCommand>>> commandsFor(String roomId) =>
      _client.get(
        ApiEndpoints.chatCommands,
        query: {'roomId': roomId},
        parse: ChatCommand.parseList,
      );
}
