/// Slash commands of the live chat, ported from
/// `web/utils/chat-parser/command.ts` (see `docs/CHAT_COMMANDS.md`).
///
/// Parsing is client side: a command expands to plain text before it is sent,
/// so the server never sees commands. The built-ins and their output must stay
/// identical to web's.
library;

import 'dart:math';

import 'package:letslive_mobile/core/chat_parser/chat_command.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';

/// Longest message the chat accepts (`CHAT_MESSAGE_MAX_LENGTH` on web,
/// `MaxLiveMessageLength` in the chat service), in UTF-16 units.
const int chatMessageMaxLength = 500;

const int _rollDefaultMax = 100;
const int _rollLargestMax = 1000000;
const int _maxSuggestions = 8;

const String _shrug = r'¯\_(ツ)_/¯';
const String _tableflip = '(╯°□°)╯︵ ┻━┻';
const String _unflip = '┬─┬ ノ( ゜-゜ノ)';

/// What a command or message typed in the chat box means.
sealed class ParsedInput {
  const ParsedInput();
}

/// Send [text] as a chat message.
class ParsedSend extends ParsedInput {
  const ParsedSend(this.text);

  final String text;
}

/// `/help`: show the command list to the sender only.
class ParsedHelp extends ParsedInput {
  const ParsedHelp();
}

/// Why a command could not run.
enum ChatInputError { unknownCommand, rollUsage }

/// A command that cannot run; shown to the sender only. [name] is the typed
/// command for [ChatInputError.unknownCommand].
class ParsedError extends ParsedInput {
  const ParsedError(this.error, {this.name = ''});

  final ChatInputError error;
  final String name;
}

/// Nothing to do (a lone `/`).
class ParsedNothing extends ParsedInput {
  const ParsedNothing();
}

/// The localized text of [error] for the sender.
String chatInputErrorText(AppLocalizations l10n, ParsedError error) =>
    switch (error.error) {
      ChatInputError.unknownCommand => l10n.chat_commands_errors_unknown(
        error.name,
      ),
      ChatInputError.rollUsage => l10n.chat_commands_errors_roll_usage,
    };

/// Where a command comes from; also the order of ties between custom ones.
enum ChatCommandSource { builtin, user, channel }

/// The `/roll` text for [value] out of [max]; web localizes it at send time.
typedef RollMessage = String Function(int value, int max);

/// A command shipped with the app.
class BuiltinChatCommand {
  const BuiltinChatCommand({
    required this.name,
    required this.usage,
    required this.description,
  });

  final String name;

  /// Shown in suggestions and `/help`, e.g. `/roll [max]`.
  final String usage;

  /// The localized one-line description.
  final String Function(AppLocalizations l10n) description;
}

/// The built-ins in web's order.
final List<BuiltinChatCommand> builtinChatCommands = [
  BuiltinChatCommand(
    name: 'shrug',
    usage: '/shrug [text]',
    description: (l10n) => l10n.chat_commands_builtin_shrug,
  ),
  BuiltinChatCommand(
    name: 'tableflip',
    usage: '/tableflip [text]',
    description: (l10n) => l10n.chat_commands_builtin_tableflip,
  ),
  BuiltinChatCommand(
    name: 'unflip',
    usage: '/unflip [text]',
    description: (l10n) => l10n.chat_commands_builtin_unflip,
  ),
  BuiltinChatCommand(
    name: 'roll',
    usage: '/roll [max]',
    description: (l10n) => l10n.chat_commands_builtin_roll,
  ),
  BuiltinChatCommand(
    name: 'help',
    usage: '/help',
    description: (l10n) => l10n.chat_commands_builtin_help,
  ),
];

/// One row of the suggestion list.
class ChatCommandSuggestion {
  const ChatCommandSuggestion({
    required this.id,
    required this.name,
    required this.description,
    required this.usage,
    required this.source,
  });

  /// `builtin:<name>` for built-ins, else the custom command's id.
  final String id;
  final String name;
  final String description;
  final String usage;
  final ChatCommandSource source;
}

/// Custom commands in tie-break order: channel before user.
List<ChatCommand> _orderedCustom(List<ChatCommand> custom) => [
  for (final scope in const [ChatCommandScope.channel, ChatCommandScope.user])
    ...custom.where((c) => c.scope == scope),
];

/// Every command that can be suggested: built-ins, then channel, then user.
List<ChatCommandSuggestion> buildChatCommandIndex(
  List<ChatCommand> custom,
  AppLocalizations l10n,
) => [
  for (final c in builtinChatCommands)
    ChatCommandSuggestion(
      id: 'builtin:${c.name}',
      name: c.name,
      description: c.description(l10n),
      usage: c.usage,
      source: ChatCommandSource.builtin,
    ),
  for (final c in _orderedCustom(custom))
    ChatCommandSuggestion(
      id: c.id,
      name: c.name,
      description: c.description.isEmpty ? c.response : c.description,
      usage: '/${c.name}',
      source: c.scope == ChatCommandScope.channel
          ? ChatCommandSource.channel
          : ChatCommandSource.user,
    ),
];

/// Suggestions for [input]: only while it is `/` plus a name with no space
/// yet; matches by prefix, at most 8.
List<ChatCommandSuggestion> filterChatCommandSuggestions(
  List<ChatCommandSuggestion> index,
  String input,
) {
  if (!input.startsWith('/')) return const [];
  final after = input.substring(1);
  if (after.contains(' ')) return const [];
  final q = after.toLowerCase();
  return index
      .where((c) => c.name.startsWith(q))
      .take(_maxSuggestions)
      .toList();
}

/// The command name typed in [input] (lowercase, no slash), or empty.
String chatCommandName(String input) {
  if (!input.startsWith('/')) return '';
  final space = input.indexOf(' ');
  return (space == -1 ? input.substring(1) : input.substring(1, space))
      .trim()
      .toLowerCase();
}

/// The `/help` text: built-ins, then the custom commands with their source.
String buildChatCommandHelpText(
  List<ChatCommand> custom,
  AppLocalizations l10n,
) {
  final lines = [
    l10n.chat_commands_help_header,
    for (final c in builtinChatCommands) '${c.usage} — ${c.description(l10n)}',
  ];
  if (custom.isNotEmpty) {
    lines.add(l10n.chat_commands_help_custom_header);
    for (final c in _orderedCustom(custom)) {
      final source = c.scope == ChatCommandScope.channel
          ? l10n.chat_commands_source_channel
          : l10n.chat_commands_source_user;
      lines.add(
        '/${c.name} — ${c.description.isEmpty ? c.response : c.description}'
        ' ($source)',
      );
    }
  }
  return lines.join('\n');
}

String _englishRoll(int value, int max) => '🎲 rolled $value (1-$max)';

final RegExp _leadingInt = RegExp(r'^[+-]?\d+');

/// Cuts [text] to [chatMessageMaxLength] UTF-16 units without splitting a
/// surrogate pair.
String _truncate(String text) {
  if (text.length <= chatMessageMaxLength) return text;
  var end = chatMessageMaxLength;
  final last = text.codeUnitAt(end - 1);
  if (last >= 0xD800 && last <= 0xDBFF) end--;
  return text.substring(0, end);
}

ParsedInput _runBuiltin(
  String name,
  String args,
  Random random,
  RollMessage rollMessage,
) {
  final trimmed = args.trim();
  String append(String face) => trimmed.isEmpty ? face : '$trimmed $face';
  switch (name) {
    case 'shrug':
      return ParsedSend(append(_shrug));
    case 'tableflip':
      return ParsedSend(append(_tableflip));
    case 'unflip':
      return ParsedSend(append(_unflip));
    case 'roll':
      var max = _rollDefaultMax;
      if (trimmed.isNotEmpty) {
        // JS parseInt: the leading digits count, the rest is ignored.
        final digits = _leadingInt.stringMatch(trimmed);
        final n = digits == null ? null : int.tryParse(digits);
        if (n == null || n < 1 || n > _rollLargestMax) {
          return const ParsedError(ChatInputError.rollUsage);
        }
        max = n;
      }
      return ParsedSend(rollMessage(random.nextInt(max) + 1, max));
    default:
      return const ParsedHelp();
  }
}

/// Turns what the user typed into what to do with it.
///
/// Plain text is sent as typed (trimmed). `/name args` runs a built-in or a
/// custom command from [commands]; when several share the name, the one the
/// user picked ([pickedId], a suggestion id) wins, otherwise built-in, then
/// channel, then user. An expanded command is cut to
/// [chatMessageMaxLength]. [random] and [rollMessage] make `/roll`
/// deterministic and localizable.
ParsedInput parseChatInput(
  String input, {
  required List<ChatCommand> commands,
  String? pickedId,
  Random? random,
  RollMessage rollMessage = _englishRoll,
}) {
  final raw = input.trim();
  if (raw.isEmpty) return const ParsedNothing();
  if (!raw.startsWith('/')) return ParsedSend(raw);

  final name = chatCommandName(raw);
  if (name.isEmpty) return const ParsedNothing();
  final space = raw.indexOf(' ');
  final args = space == -1 ? '' : raw.substring(space + 1);

  final matches = _orderedCustom(commands).where((c) => c.name == name);
  final picked = matches.where((c) => c.id == pickedId).firstOrNull;
  final result = picked != null
      ? ParsedSend(picked.response)
      : builtinChatCommands.any((c) => c.name == name)
      ? _runBuiltin(name, args, random ?? Random(), rollMessage)
      : matches.isNotEmpty
      ? ParsedSend(matches.first.response)
      : ParsedError(ChatInputError.unknownCommand, name: name);
  return result is ParsedSend ? ParsedSend(_truncate(result.text)) : result;
}
