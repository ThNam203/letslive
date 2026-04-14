enum ChatCommandScope { user, channel }

ChatCommandScope _scopeFromString(String s) {
  return s == 'channel' ? ChatCommandScope.channel : ChatCommandScope.user;
}

String chatCommandScopeToString(ChatCommandScope s) =>
    s == ChatCommandScope.channel ? 'channel' : 'user';

class ChatCommand {
  final String id;
  final ChatCommandScope scope;
  final String ownerId;
  final String name;
  final String response;
  final String description;

  const ChatCommand({
    required this.id,
    required this.scope,
    required this.ownerId,
    required this.name,
    required this.response,
    required this.description,
  });

  factory ChatCommand.fromJson(Map<String, dynamic> json) {
    return ChatCommand(
      id: json['id'] as String? ?? json['_id'] as String? ?? '',
      scope: _scopeFromString(json['scope'] as String? ?? 'user'),
      ownerId: json['ownerId'] as String? ?? '',
      name: json['name'] as String? ?? '',
      response: json['response'] as String? ?? '',
      description: json['description'] as String? ?? '',
    );
  }
}

class MyChatCommands {
  final List<ChatCommand> user;
  final List<ChatCommand> channel;

  const MyChatCommands({required this.user, required this.channel});

  factory MyChatCommands.fromJson(Map<String, dynamic> json) {
    final user = (json['user'] as List<dynamic>? ?? [])
        .map((e) => ChatCommand.fromJson(e as Map<String, dynamic>))
        .toList();
    final channel = (json['channel'] as List<dynamic>? ?? [])
        .map((e) => ChatCommand.fromJson(e as Map<String, dynamic>))
        .toList();
    return MyChatCommands(user: user, channel: channel);
  }
}
