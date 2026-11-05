import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/messages/data/dm_models.dart';
import 'package:letslive_mobile/features/messages/domain/dm_rules.dart';

/// The `/conversations` endpoints, mirroring `web/lib/api/dm.ts`. Failures
/// are values ([ApiResponse.success] false), never thrown.
class MessageRepository {
  const MessageRepository(this._client);

  final ApiClient _client;

  static const String _base = ApiEndpoints.conversations;

  /// `GET /conversations`: the user's conversations, newest activity first,
  /// [dmConversationsPageSize] per [page] (0-based).
  Future<ApiResponse<List<Conversation>>> list({
    int page = 0,
    int limit = dmConversationsPageSize,
  }) => _client.get(
    _base,
    query: {'page': page, 'limit': limit},
    parse: Conversation.parseList,
  );

  /// `GET /conversations/<id>`; 404 or 403 when it does not exist or the
  /// user is not a member.
  Future<ApiResponse<Conversation>> byId(String id) =>
      _client.get('$_base/$id', parse: Conversation.parse);

  /// `POST /conversations`. Creating a DM that already exists answers the
  /// existing one.
  Future<ApiResponse<Conversation>> create({
    required ConversationType type,
    required List<String> participantIds,
    String? name,
  }) => _client.post(
    _base,
    body: {'type': type.name, 'participantIds': participantIds, 'name': ?name},
    parse: Conversation.parse,
  );

  /// `PUT /conversations/<id>`: renames a group or changes its picture.
  Future<ApiResponse<Conversation>> update(
    String id, {
    String? name,
    String? avatarUrl,
  }) => _client.put(
    '$_base/$id',
    body: {'name': ?name, 'avatarUrl': ?avatarUrl},
    parse: Conversation.parse,
  );

  /// `DELETE /conversations/<id>`: the user leaves it.
  Future<ApiResponse<Object?>> leave(String id) =>
      _client.delete('$_base/$id', parse: (data) => data);

  /// `POST /conversations/<id>/participants`.
  Future<ApiResponse<Conversation>> addParticipant(
    String id, {
    required String userId,
    required String username,
    String? profilePicture,
  }) => _client.post(
    '$_base/$id/participants',
    body: {
      'userId': userId,
      'username': username,
      'profilePicture': ?profilePicture,
    },
    parse: Conversation.parse,
  );

  /// `DELETE /conversations/<id>/participants/<userId>`.
  Future<ApiResponse<Conversation>> removeParticipant(
    String id,
    String userId,
  ) => _client.delete(
    '$_base/$id/participants/$userId',
    parse: Conversation.parse,
  );

  /// `GET /conversations/<id>/messages`: up to [limit] messages sent before
  /// the message [before] (the newest when null), oldest first.
  Future<ApiResponse<List<DmMessage>>> messages(
    String id, {
    String? before,
    int limit = dmMessagesPageSize,
  }) => _client.get(
    '$_base/$id/messages',
    query: {'limit': limit, 'before': ?before},
    parse: DmMessage.parseList,
  );

  /// `POST /conversations/<id>/messages`: sends a text message. The server
  /// also pushes it to every member, the sender included.
  Future<ApiResponse<DmMessage>> send(String id, String text) => _client.post(
    '$_base/$id/messages',
    body: {'text': text, 'type': DmMessageType.text.name},
    parse: DmMessage.parse,
  );

  /// `PATCH /conversations/<id>/messages/<messageId>`: edits an own message.
  Future<ApiResponse<DmMessage>> edit(
    String id,
    String messageId,
    String text,
  ) => _client.patch(
    '$_base/$id/messages/$messageId',
    body: {'text': text},
    parse: DmMessage.parse,
  );

  /// `DELETE /conversations/<id>/messages/<messageId>`: deletes an own
  /// message (the server keeps it as "deleted").
  Future<ApiResponse<Object?>> delete(String id, String messageId) =>
      _client.delete('$_base/$id/messages/$messageId', parse: (data) => data);

  /// `POST /conversations/<id>/read`: moves the read marker to [messageId],
  /// or to the newest message when null.
  Future<ApiResponse<Object?>> markRead(String id, {String? messageId}) =>
      _client.post(
        '$_base/$id/read',
        body: {'messageId': ?messageId},
        parse: (data) => data,
      );

  /// `POST /conversations/<id>/typing`: tells the others the user started or
  /// stopped typing (the server only relays it).
  Future<ApiResponse<Object?>> typing(String id, {required bool isTyping}) =>
      _client.post(
        '$_base/$id/typing',
        body: {'state': isTyping ? 'start' : 'stop'},
        parse: (data) => data,
      );

  /// `GET /conversations/unread-counts`: unread messages per conversation id
  /// (conversations without unread messages are absent).
  Future<ApiResponse<Map<String, int>>> unreadCounts() => _client.get(
    '$_base/unread-counts',
    parse: (data) {
      if (data is! Map<String, Object?>) {
        throw FormatException('unread-counts: expected an object, got $data');
      }
      return {
        for (final MapEntry(:key, :value) in data.entries)
          key: value is int
              ? value
              : throw FormatException('unread-counts.$key: got $value'),
      };
    },
  );
}
