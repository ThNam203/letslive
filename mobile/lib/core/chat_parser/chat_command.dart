import 'package:letslive_mobile/core/network/json_fields.dart';

/// Who a custom [ChatCommand] belongs to (`web/types/chat-command.ts`).
enum ChatCommandScope {
  /// Created by a user, usable by that user in any channel.
  user,

  /// Created by a streamer, usable by every viewer of that channel.
  channel,
}

/// A custom slash command: typing `/<name>` sends [response]. Parsed
/// strictly after `dto.ChatCommandResponse` in `backend/chat`.
class ChatCommand {
  const ChatCommand({
    required this.id,
    required this.scope,
    required this.ownerId,
    required this.name,
    required this.response,
    required this.description,
  });

  factory ChatCommand.fromJson(Map<String, Object?> json) {
    final f = JsonFields('ChatCommand', json);
    final scope = f.string('scope');
    return ChatCommand(
      id: f.string('id'),
      scope: ChatCommandScope.values.firstWhere(
        (s) => s.name == scope,
        orElse: () => throw FormatException('ChatCommand.scope: "$scope"'),
      ),
      ownerId: f.string('ownerId'),
      name: f.string('name'),
      response: f.string('response'),
      description: f.optString('description') ?? '',
    );
  }

  /// `ParseData<List<ChatCommand>>` for `ApiClient`.
  static List<ChatCommand> parseList(Object? data) =>
      parseJsonList(data, 'ChatCommand', ChatCommand.fromJson);

  /// Server-assigned id; picks between commands that share a name.
  final String id;

  /// Whose command it is: a user's own or a channel's.
  final ChatCommandScope scope;

  /// The user who created it (for a channel command, the streamer).
  final String ownerId;

  /// Lowercase `[a-z0-9_-]{1,32}`, typed after the slash.
  final String name;

  /// The text sent when the command runs.
  final String response;

  /// Short help text; empty when the owner wrote none.
  final String description;
}
