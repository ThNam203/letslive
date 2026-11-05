import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/chat_parser/chat_command.dart';
import 'package:letslive_mobile/core/chat_parser/command.dart';
import 'package:letslive_mobile/core/chat_parser/emote.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/watch/chat/chat_providers.dart';
import 'package:letslive_mobile/features/watch/chat/data/chat_message.dart';
import 'package:letslive_mobile/features/watch/chat/presentation/emote_picker.dart';

/// The message box under the chat list (web `ChatPanel` form): text field,
/// emote button, send button. The `/command` and `:emote` suggestions are
/// published to [hints], which the panel draws over the list so they can
/// never be taller than the space the list has.
///
/// Submitting runs [parseChatInput]: plain text and expanded commands go to
/// the server; `/help` and command errors become notices only the sender
/// sees. The text is cleared only after the server accepted the message, and
/// only if the field still holds what was sent; a failure keeps it and shows
/// the server's reason with a retry that resends the exact failed text, so a
/// message is never lost or replaced by a re-parse (`/roll` does not re-roll).
class ChatInput extends ConsumerStatefulWidget {
  const ChatInput({required this.roomId, required this.hints, super.key});

  /// The room (streamer user id) the message goes to.
  final String roomId;

  /// Where the suggestion list or emote strip is published (null clears).
  final ValueNotifier<Widget?> hints;

  @override
  ConsumerState<ChatInput> createState() => _ChatInputState();
}

/// A send the server refused: `text` is what was posted, `raw` what the
/// field held then.
typedef _FailedSend = ({
  String text,
  String raw,
  ApiResponse<ChatMessage> response,
});

class _ChatInputState extends ConsumerState<ChatInput> {
  final _text = TextEditingController();
  // The suggestion the user tapped; decides between commands sharing a name.
  ChatCommandSuggestion? _picked;
  bool _sending = false;
  _FailedSend? _failed;

  @override
  void dispose() {
    _text.dispose();
    super.dispose();
  }

  List<ChatCommand> get _commands =>
      ref.read(chatCommandsProvider(widget.roomId)).value ?? const [];

  /// Publishes the command suggestions for [value], else the emote
  /// suggestions for the shortcode being typed, else nothing.
  void _updateHints(String value) {
    final index = buildChatCommandIndex(
      _commands,
      AppLocalizations.of(context),
    );
    final commands = filterChatCommandSuggestions(index, value);
    final cursor = _text.selection.baseOffset;
    final before = cursor < 0 || cursor > value.length
        ? value
        : value.substring(0, cursor);
    final query = activeEmoteQuery(before);
    final matches = query == null ? const <Emote>[] : searchEmotes(query);
    widget.hints.value = commands.isNotEmpty
        ? _SuggestionList(suggestions: commands, onPick: _pickSuggestion)
        : matches.isNotEmpty
        ? _EmoteStrip(emotes: matches, onPick: _pickEmote)
        : null;
  }

  void _onChanged(String value) {
    final picked = _picked;
    if (picked != null && chatCommandName(value) != picked.name) _picked = null;
    _updateHints(value);
    if (_failed != null) setState(() => _failed = null);
  }

  void _pickSuggestion(ChatCommandSuggestion s) {
    _text.value = TextEditingValue(
      text: '/${s.name} ',
      selection: TextSelection.collapsed(offset: s.name.length + 2),
    );
    _picked = s;
    widget.hints.value = null;
  }

  void _pickEmote(Emote emote) {
    final result = insertEmote(_text.text, _text.selection.baseOffset, emote);
    _text.value = TextEditingValue(
      text: result.text,
      selection: TextSelection.collapsed(offset: result.cursor),
    );
    _onChanged(result.text);
  }

  void _clear() {
    _text.clear();
    _picked = null;
    widget.hints.value = null;
    setState(() => _failed = null);
  }

  Future<void> _openEmotePicker() async {
    final emote = await showEmotePicker(context);
    if (emote == null || !mounted) return;
    _pickEmote(emote);
  }

  /// The send button / keyboard action. While a send is in flight nothing
  /// happens: the text simply stays in the field until the send is done.
  Future<void> _submit() async {
    if (_sending) return;
    final l10n = AppLocalizations.of(context);
    final chat = ref.read(chatControllerProvider(widget.roomId).notifier);
    final commands = _commands;
    final raw = _text.text;
    final parsed = parseChatInput(
      raw,
      commands: commands,
      pickedId: _picked?.id,
      rollMessage: (value, max) => l10n.chat_commands_roll_message(value, max),
    );
    switch (parsed) {
      case ParsedNothing():
        if (raw.trim().startsWith('/')) _clear();
      case ParsedHelp():
        chat.addNotice(buildChatCommandHelpText(commands, l10n));
        _clear();
      case ParsedError():
        chat.addNotice(chatInputErrorText(l10n, parsed));
        _clear();
      case ParsedSend(:final text):
        await _send(text, raw);
    }
  }

  /// Posts [text]; [raw] is what the field held, to know whether it may be
  /// cleared afterwards.
  Future<void> _send(String text, String raw) async {
    setState(() {
      _sending = true;
      _failed = null;
    });
    final response = await ref
        .read(chatControllerProvider(widget.roomId).notifier)
        .send(text);
    if (!mounted) return;
    setState(() => _sending = false);
    if (!response.success) {
      setState(() => _failed = (text: text, raw: raw, response: response));
    } else if (_text.text == raw) {
      _clear();
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final failed = _failed;
    // Loaded here so the commands are ready by the time a "/" is typed.
    ref.watch(chatCommandsProvider(widget.roomId));
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        if (failed != null)
          _SendFailed(
            reason: apiErrorMessage(l10n, failed.response),
            onRetry: () => _send(failed.text, failed.raw),
          ),
        Padding(
          padding: const EdgeInsets.fromLTRB(8, 4, 8, 8),
          child: Row(
            children: [
              IconButton(
                tooltip: l10n.chat_emotes_action,
                icon: const Icon(Icons.emoji_emotions_outlined),
                onPressed: _openEmotePicker,
              ),
              Expanded(
                child: TextField(
                  controller: _text,
                  inputFormatters: [_serverLengthLimit],
                  textInputAction: TextInputAction.send,
                  decoration: InputDecoration(
                    hintText: l10n.chat_placeholder_typing,
                    isDense: true,
                    border: const OutlineInputBorder(),
                  ),
                  onChanged: _onChanged,
                  // Overridden so a send in flight does not also close the
                  // keyboard; the text stays visible in the field.
                  onEditingComplete: () => unawaited(_submit()),
                ),
              ),
              IconButton(
                tooltip: l10n.chat_send_action,
                icon: const Icon(Icons.send),
                onPressed: _sending ? null : _submit,
              ),
            ],
          ),
        ),
      ],
    );
  }
}

/// Caps the field at the server's limit, which counts UTF-16 units (an
/// emoji is 2), like web's `maxLength`. Pasted text is cut, not refused.
final TextInputFormatter _serverLengthLimit = TextInputFormatter.withFunction((
  old,
  value,
) {
  if (value.text.length <= chatMessageMaxLength) return value;
  final cut = truncateChatMessage(value.text);
  return TextEditingValue(
    text: cut,
    selection: TextSelection.collapsed(offset: cut.length),
  );
});

/// Why the message was not sent, with a retry; the text is still in the box.
class _SendFailed extends StatelessWidget {
  const _SendFailed({required this.reason, required this.onRetry});

  final String reason;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 12),
      child: Wrap(
        crossAxisAlignment: WrapCrossAlignment.center,
        spacing: 8,
        children: [
          Text(
            '${l10n.chat_send_failed} $reason',
            style: TextStyle(color: theme.colorScheme.error),
          ),
          TextButton(onPressed: onRetry, child: Text(l10n.chat_send_retry)),
        ],
      ),
    );
  }
}

/// Emotes matching the shortcode being typed, as one scrollable row.
class _EmoteStrip extends StatelessWidget {
  const _EmoteStrip({required this.emotes, required this.onPick});

  final List<Emote> emotes;
  final ValueChanged<Emote> onPick;

  @override
  Widget build(BuildContext context) => Material(
    elevation: 2,
    child: SizedBox(
      height: 52,
      child: ListView.builder(
        scrollDirection: Axis.horizontal,
        itemCount: emotes.length,
        itemBuilder: (_, i) => Tooltip(
          message: emotes[i].shortcode,
          child: InkWell(
            onTap: () => onPick(emotes[i]),
            child: SizedBox(
              width: 48,
              child: Center(
                child: Text(
                  emotes[i].emoji,
                  style: const TextStyle(fontSize: 24),
                ),
              ),
            ),
          ),
        ),
      ),
    ),
  );
}

/// The matching commands, each with its usage, description and source. The
/// panel limits its height; it scrolls when the space is short.
class _SuggestionList extends StatelessWidget {
  const _SuggestionList({required this.suggestions, required this.onPick});

  final List<ChatCommandSuggestion> suggestions;
  final ValueChanged<ChatCommandSuggestion> onPick;

  String _source(AppLocalizations l10n, ChatCommandSource source) =>
      switch (source) {
        ChatCommandSource.builtin => l10n.chat_commands_source_builtin,
        ChatCommandSource.user => l10n.chat_commands_source_user,
        ChatCommandSource.channel => l10n.chat_commands_source_channel,
      };

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Material(
      elevation: 2,
      // At most 8 rows, so no need for a lazy list.
      child: SingleChildScrollView(
        child: Column(
          children: [
            for (final s in suggestions)
              ListTile(
                key: ValueKey('suggestion-${s.id}'),
                dense: true,
                title: Text(
                  s.usage,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontFamily: 'monospace'),
                ),
                subtitle: Text(
                  '${_source(l10n, s.source)} - ${s.description}',
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
                onTap: () => onPick(s),
              ),
          ],
        ),
      ),
    );
  }
}
