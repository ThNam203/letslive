import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/network/json_fields.dart';

/// Whose chat the command is available in: the creator's own (`user`, in any
/// channel) or everyone chatting in their channel (`channel`).
enum ChatCommandScope {
  user,
  channel;

  static ChatCommandScope parse(String value) => switch (value) {
    'user' => user,
    'channel' => channel,
    _ => throw FormatException('ChatCommand.scope: unknown "$value"'),
  };
}

/// One custom chat command (web `ChatCommand`).
class ChatCommandEntry {
  const ChatCommandEntry({
    required this.id,
    required this.scope,
    required this.name,
    required this.response,
    required this.description,
  });

  factory ChatCommandEntry.fromJson(Map<String, Object?> json) {
    final f = JsonFields('ChatCommand', json);
    return ChatCommandEntry(
      id: f.string('id'),
      scope: ChatCommandScope.parse(f.string('scope')),
      name: f.string('name'),
      response: f.string('response'),
      description: f.optString('description') ?? '',
    );
  }

  /// `ParseData<ChatCommandEntry>`.
  static ChatCommandEntry parse(Object? data) =>
      parseJsonObject(data, 'ChatCommand', ChatCommandEntry.fromJson);

  final String id;
  final ChatCommandScope scope;

  /// Without the leading slash.
  final String name;
  final String response;
  final String description;
}

/// The creator's own commands, split by scope (web `MyChatCommands`).
class MyChatCommands {
  const MyChatCommands({required this.user, required this.channel});

  /// `ParseData<MyChatCommands>`. Go sends `null` for an empty list.
  static MyChatCommands parse(Object? data) {
    if (data is! Map<String, Object?>) {
      throw FormatException('MyChatCommands: expected an object');
    }
    List<ChatCommandEntry> list(String key) => data[key] == null
        ? const []
        : parseJsonList(data[key], 'ChatCommand', ChatCommandEntry.fromJson);
    return MyChatCommands(user: list('user'), channel: list('channel'));
  }

  final List<ChatCommandEntry> user;
  final List<ChatCommandEntry> channel;

  /// The commands of [scope].
  List<ChatCommandEntry> of(ChatCommandScope scope) =>
      scope == ChatCommandScope.user ? user : channel;
}

/// The chat-command admin calls (`web/lib/api/chat-command.ts`). Failures
/// are values, never thrown.
class ChatCommandAdminRepository {
  const ChatCommandAdminRepository(this._client);

  final ApiClient _client;

  /// `GET /chat-commands/mine`.
  Future<ApiResponse<MyChatCommands>> mine() => _client.get(
    '${ApiEndpoints.chatCommands}/mine',
    parse: MyChatCommands.parse,
  );

  /// `POST /chat-commands`. The name must be unique per owner and scope
  /// (the server answers a conflict otherwise).
  Future<ApiResponse<ChatCommandEntry>> create({
    required ChatCommandScope scope,
    required String name,
    required String response,
    required String description,
  }) => _client.post(
    ApiEndpoints.chatCommands,
    body: {
      'scope': scope.name,
      'name': name,
      'response': response,
      'description': description,
    },
    parse: ChatCommandEntry.parse,
  );

  /// `PATCH /chat-commands/<id>`.
  Future<ApiResponse<ChatCommandEntry>> update(
    String id, {
    required String name,
    required String response,
    required String description,
  }) => _client.patch(
    '${ApiEndpoints.chatCommands}/$id',
    body: {'name': name, 'response': response, 'description': description},
    parse: ChatCommandEntry.parse,
  );

  /// `DELETE /chat-commands/<id>`.
  Future<ApiResponse<Object?>> delete(String id) =>
      _client.delete('${ApiEndpoints.chatCommands}/$id', parse: (data) => data);
}
