import 'dart:math';

import '../../l10n/app_localizations.dart';
import '../../models/chat_command.dart';

/// Outcome of parsing a slash chat command.
sealed class ChatCommandResult {
  const ChatCommandResult();
}

/// Expand to message text and send via the normal chat path.
class ChatCommandSend extends ChatCommandResult {
  final String text;
  const ChatCommandSend(this.text);
}

/// Show the local-only help message.
class ChatCommandHelp extends ChatCommandResult {
  const ChatCommandHelp();
}

/// Show a local-only error message to the sender.
class ChatCommandError extends ChatCommandResult {
  final String message;
  const ChatCommandError(this.message);
}

/// Empty / no-op (e.g. `/me` with no text).
class ChatCommandNoop extends ChatCommandResult {
  const ChatCommandNoop();
}

/// Where a suggestion came from.
enum ChatCommandSource { builtin, user, channel }

class ChatCommandSuggestion {
  final String name;
  final String description;
  final String usage;
  final ChatCommandSource source;

  const ChatCommandSuggestion({
    required this.name,
    required this.description,
    required this.usage,
    required this.source,
  });
}

class BuiltinChatCommand {
  final String name;
  final String usage;
  final String Function(AppLocalizations l10n) describe;
  final ChatCommandResult Function(String args, AppLocalizations l10n) run;

  const BuiltinChatCommand({
    required this.name,
    required this.usage,
    required this.describe,
    required this.run,
  });
}

const _shrug = '¯\\_(ツ)_/¯';
const _tableflip = '(╯°□°)╯︵ ┻━┻';
const _unflip = '┬─┬ ノ( ゜-゜ノ)';

final List<BuiltinChatCommand> builtinChatCommands = [
  BuiltinChatCommand(
    name: 'me',
    usage: '/me <text>',
    describe: (l) => l.chatCommandsBuiltinMe,
    run: (args, l) {
      final t = args.trim();
      if (t.isEmpty) return const ChatCommandNoop();
      return ChatCommandSend('_${t}_');
    },
  ),
  BuiltinChatCommand(
    name: 'shrug',
    usage: '/shrug [text]',
    describe: (l) => l.chatCommandsBuiltinShrug,
    run: (args, l) {
      final t = args.trim();
      return ChatCommandSend(t.isEmpty ? _shrug : '$t $_shrug');
    },
  ),
  BuiltinChatCommand(
    name: 'tableflip',
    usage: '/tableflip [text]',
    describe: (l) => l.chatCommandsBuiltinTableflip,
    run: (args, l) {
      final t = args.trim();
      return ChatCommandSend(t.isEmpty ? _tableflip : '$t $_tableflip');
    },
  ),
  BuiltinChatCommand(
    name: 'unflip',
    usage: '/unflip [text]',
    describe: (l) => l.chatCommandsBuiltinUnflip,
    run: (args, l) {
      final t = args.trim();
      return ChatCommandSend(t.isEmpty ? _unflip : '$t $_unflip');
    },
  ),
  BuiltinChatCommand(
    name: 'roll',
    usage: '/roll [max]',
    describe: (l) => l.chatCommandsBuiltinRoll,
    run: (args, l) {
      final t = args.trim();
      var max = 100;
      if (t.isNotEmpty) {
        final n = int.tryParse(t);
        if (n == null || n < 1 || n > 1000000) {
          return ChatCommandError(l.chatCommandsErrorRollUsage);
        }
        max = n;
      }
      final value = Random().nextInt(max) + 1;
      return ChatCommandSend(l.chatCommandsRollMessage(value, max));
    },
  ),
  BuiltinChatCommand(
    name: 'help',
    usage: '/help',
    describe: (l) => l.chatCommandsBuiltinHelp,
    run: (args, l) => const ChatCommandHelp(),
  ),
];

final Map<String, BuiltinChatCommand> _builtinMap = {
  for (final c in builtinChatCommands) c.name: c,
};

/// Build a unified list of suggestions (built-ins first, then customs).
List<ChatCommandSuggestion> buildChatCommandIndex(
  List<ChatCommand> custom,
  AppLocalizations l10n,
) {
  final seen = <String>{};
  final out = <ChatCommandSuggestion>[];
  for (final c in builtinChatCommands) {
    if (!seen.add(c.name)) continue;
    out.add(
      ChatCommandSuggestion(
        name: c.name,
        description: c.describe(l10n),
        usage: c.usage,
        source: ChatCommandSource.builtin,
      ),
    );
  }
  for (final c in custom) {
    if (!seen.add(c.name)) continue;
    out.add(
      ChatCommandSuggestion(
        name: c.name,
        description: c.description.isNotEmpty ? c.description : c.response,
        usage: '/${c.name}',
        source: c.scope == ChatCommandScope.user
            ? ChatCommandSource.user
            : ChatCommandSource.channel,
      ),
    );
  }
  return out;
}

/// Filter the index by the current input prefix. Returns at most 8 entries.
List<ChatCommandSuggestion> filterChatCommandSuggestions(
  List<ChatCommandSuggestion> index,
  String input,
) {
  if (!input.startsWith('/')) return const [];
  final after = input.substring(1);
  if (after.contains(' ')) return const [];
  final q = after.toLowerCase();
  return index.where((s) => s.name.startsWith(q)).take(8).toList();
}

/// Parse the input. Returns null when [input] is not a slash command.
ChatCommandResult? parseChatCommand(
  String input,
  List<ChatCommand> custom,
  AppLocalizations l10n,
) {
  if (!input.startsWith('/')) return null;
  final spaceIdx = input.indexOf(' ');
  final name =
      (spaceIdx == -1 ? input.substring(1) : input.substring(1, spaceIdx))
          .trim()
          .toLowerCase();
  final args = spaceIdx == -1 ? '' : input.substring(spaceIdx + 1);
  if (name.isEmpty) return null;

  final builtin = _builtinMap[name];
  if (builtin != null) return builtin.run(args, l10n);

  for (final c in custom) {
    if (c.name == name) return ChatCommandSend(c.response);
  }

  return ChatCommandError(l10n.chatCommandsErrorUnknown(name));
}

/// Compose the local-only `/help` text.
String buildChatCommandHelpText(
  List<ChatCommand> custom,
  AppLocalizations l10n,
) {
  final lines = <String>[
    l10n.chatCommandsHelpHeader,
    for (final c in builtinChatCommands) '${c.usage} — ${c.describe(l10n)}',
  ];
  if (custom.isNotEmpty) {
    lines.add(l10n.chatCommandsHelpCustomHeader);
    for (final c in custom) {
      lines.add(
        '/${c.name} — ${c.description.isNotEmpty ? c.description : c.response}',
      );
    }
  }
  return lines.join('\n');
}
