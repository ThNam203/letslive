import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/chat_parser/chat_command.dart';
import 'package:letslive_mobile/core/chat_parser/command.dart';
import 'package:letslive_mobile/core/chat_parser/emote.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/watch/chat/chat_providers.dart';
import 'package:letslive_mobile/features/watch/chat/presentation/emote_picker.dart';

/// The message box under the chat list (web `ChatPanel` form): text field,
/// emote button, send button, and the `/command` suggestions above.
///
/// Submitting runs [parseChatInput]: plain text and expanded commands go to
/// the server; `/help` and command errors become notices only the sender
/// sees. The text is cleared only after the server accepted the message; a
/// failure keeps it and shows a retry, so a message is never lost.
class ChatInput extends ConsumerStatefulWidget {
  const ChatInput({required this.roomId, super.key});

  /// The room (streamer user id) the message goes to.
  final String roomId;

  @override
  ConsumerState<ChatInput> createState() => _ChatInputState();
}

class _ChatInputState extends ConsumerState<ChatInput> {
  final _text = TextEditingController();
  List<ChatCommandSuggestion> _suggestions = const [];
  // The suggestion the user tapped; decides between commands sharing a name.
  ChatCommandSuggestion? _picked;
  bool _sending = false;
  bool _failed = false;

  @override
  void dispose() {
    _text.dispose();
    super.dispose();
  }

  List<ChatCommand> get _commands =>
      ref.read(chatCommandsProvider(widget.roomId)).value ?? const [];

  void _onChanged(String value) {
    final picked = _picked;
    if (picked != null && chatCommandName(value) != picked.name) _picked = null;
    final index = buildChatCommandIndex(
      _commands,
      AppLocalizations.of(context),
    );
    setState(() {
      _suggestions = filterChatCommandSuggestions(index, value);
      _failed = false;
    });
  }

  void _pickSuggestion(ChatCommandSuggestion s) {
    _text.value = TextEditingValue(
      text: '/${s.name} ',
      selection: TextSelection.collapsed(offset: s.name.length + 2),
    );
    setState(() {
      _picked = s;
      _suggestions = const [];
    });
  }

  void _clear() {
    _text.clear();
    setState(() {
      _picked = null;
      _suggestions = const [];
      _failed = false;
    });
  }

  Future<void> _insertEmote() async {
    final emote = await showEmotePicker(context);
    if (emote == null || !mounted) return;
    final result = insertEmote(_text.text, _text.selection.baseOffset, emote);
    _text.value = TextEditingValue(
      text: result.text,
      selection: TextSelection.collapsed(offset: result.cursor),
    );
    _onChanged(result.text);
  }

  Future<void> _submit() async {
    if (_sending) return;
    final l10n = AppLocalizations.of(context);
    final chat = ref.read(chatControllerProvider(widget.roomId).notifier);
    final commands = _commands;
    final parsed = parseChatInput(
      _text.text,
      commands: commands,
      pickedId: _picked?.id,
      rollMessage: (value, max) => l10n.chat_commands_roll_message(value, max),
    );
    switch (parsed) {
      case ParsedNothing():
        if (_text.text.trim().startsWith('/')) _clear();
      case ParsedHelp():
        chat.addNotice(buildChatCommandHelpText(commands, l10n));
        _clear();
      case ParsedError():
        chat.addNotice(chatInputErrorText(l10n, parsed));
        _clear();
      case ParsedSend(:final text):
        setState(() {
          _sending = true;
          _failed = false;
        });
        final response = await chat.send(text);
        if (!mounted) return;
        setState(() => _sending = false);
        if (response.success) {
          _clear();
        } else {
          setState(() => _failed = true);
        }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    // Loaded here so the commands are ready by the time a "/" is typed.
    ref.watch(chatCommandsProvider(widget.roomId));
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        if (_suggestions.isNotEmpty)
          _SuggestionList(suggestions: _suggestions, onPick: _pickSuggestion),
        if (_failed) _SendFailed(onRetry: _submit),
        Padding(
          padding: const EdgeInsets.fromLTRB(8, 4, 8, 8),
          child: Row(
            children: [
              IconButton(
                tooltip: l10n.chat_emotes_action,
                icon: const Icon(Icons.emoji_emotions_outlined),
                onPressed: _insertEmote,
              ),
              Expanded(
                child: TextField(
                  controller: _text,
                  inputFormatters: [
                    LengthLimitingTextInputFormatter(chatMessageMaxLength),
                  ],
                  textInputAction: TextInputAction.send,
                  decoration: InputDecoration(
                    hintText: l10n.chat_placeholder_typing,
                    isDense: true,
                    border: const OutlineInputBorder(),
                  ),
                  onChanged: _onChanged,
                  onSubmitted: (_) => _submit(),
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

/// "Message not sent" with a retry; the text is still in the box.
class _SendFailed extends StatelessWidget {
  const _SendFailed({required this.onRetry});

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
            l10n.chat_send_failed,
            style: TextStyle(color: theme.colorScheme.error),
          ),
          TextButton(onPressed: onRetry, child: Text(l10n.chat_send_retry)),
        ],
      ),
    );
  }
}

/// The matching commands, each with its usage, description and source.
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
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxHeight: 220),
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
      ),
    );
  }
}
