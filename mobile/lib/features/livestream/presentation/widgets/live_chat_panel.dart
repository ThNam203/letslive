import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../../core/chat_parser/command.dart';
import '../../../../core/network/websocket_service.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../../models/chat_command.dart';
import '../../../../models/chat_message.dart';
import '../../../../providers.dart';
import '../../../../shared/widgets/chat_command_suggestions.dart';
import '../../../../shared/widgets/emote_picker.dart';
import 'chat_bubble.dart';
import 'chat_line.dart';

/// The live chat side-panel: history fetch, WebSocket lifecycle, slash
/// chat-command parsing, autocomplete, and message rendering.
///
/// Self-contained — the parent only needs to provide the room id and toggle
/// visibility.
class LiveChatPanel extends ConsumerStatefulWidget {
  final String roomId;

  const LiveChatPanel({super.key, required this.roomId});

  @override
  ConsumerState<LiveChatPanel> createState() => _LiveChatPanelState();
}

class _LiveChatPanelState extends ConsumerState<LiveChatPanel> {
  final TextEditingController _chatController = TextEditingController();
  final ScrollController _scrollController = ScrollController();
  LiveChatService? _chatService;
  StreamSubscription<ChatMessage>? _chatSubscription;

  final List<ChatLine> _lines = [];
  List<ChatCommand> _customChatCommands = const [];
  List<ChatCommandSuggestion> _commandSuggestions = const [];

  @override
  void initState() {
    super.initState();
    _initChat();
    _chatController.addListener(_onInputChanged);
  }

  @override
  void dispose() {
    _chatController.removeListener(_onInputChanged);
    _chatSubscription?.cancel();
    _chatService?.dispose();
    _chatController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  void _initChat() {
    final currentUser = ref.read(currentUserProvider);
    if (currentUser == null) return;

    final chatRepo = ref.read(chatRepositoryProvider);
    chatRepo.getMessages(roomId: widget.roomId).then((response) {
      if (mounted && response.success && response.data != null) {
        setState(() {
          _lines
            ..clear()
            ..addAll(response.data!.map((m) => RemoteLine(m)));
        });
        _scrollToBottom();
      }
    });

    final cmdRepo = ref.read(chatCommandRepositoryProvider);
    cmdRepo.listForRoom(widget.roomId).then((response) {
      if (mounted && response.success && response.data != null) {
        setState(() => _customChatCommands = response.data!);
      }
    });

    _chatService = LiveChatService();
    _chatSubscription = _chatService!.messages.listen((message) {
      if (mounted) {
        setState(() => _lines.add(RemoteLine(message)));
        _scrollToBottom();
      }
    });
    _chatService!.connect(
      roomId: widget.roomId,
      userId: currentUser.id,
      username: currentUser.username,
    );
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollController.hasClients) {
        _scrollController.animateTo(
          _scrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOut,
        );
      }
    });
  }

  void _onInputChanged() {
    final l10n = AppLocalizations.of(context);
    final index = buildChatCommandIndex(_customChatCommands, l10n);
    final next = filterChatCommandSuggestions(index, _chatController.text);
    if (!_suggestionsEqual(next, _commandSuggestions)) {
      setState(() => _commandSuggestions = next);
    }
  }

  bool _suggestionsEqual(
    List<ChatCommandSuggestion> a,
    List<ChatCommandSuggestion> b,
  ) {
    if (a.length != b.length) return false;
    for (var i = 0; i < a.length; i++) {
      if (a[i].name != b[i].name) return false;
    }
    return true;
  }

  void _sendMessage() {
    final raw = _chatController.text.trim();
    if (raw.isEmpty || _chatService == null) return;
    final l10n = AppLocalizations.of(context);

    if (raw.startsWith('/')) {
      final result = parseChatCommand(raw, _customChatCommands, l10n);
      _chatController.clear();
      setState(() => _commandSuggestions = const []);
      if (result == null || result is ChatCommandNoop) return;
      if (result is ChatCommandError) {
        setState(() => _lines.add(SystemLine(result.message)));
        _scrollToBottom();
        return;
      }
      if (result is ChatCommandHelp) {
        setState(
          () => _lines.add(
            SystemLine(buildChatCommandHelpText(_customChatCommands, l10n)),
          ),
        );
        _scrollToBottom();
        return;
      }
      if (result is ChatCommandSend) {
        _chatService!.sendMessage(result.text);
        return;
      }
      return;
    }

    _chatService!.sendMessage(raw);
    _chatController.clear();
    setState(() => _commandSuggestions = const []);
  }

  void _applySuggestion(ChatCommandSuggestion s) {
    final next = '/${s.name} ';
    _chatController.value = TextEditingValue(
      text: next,
      selection: TextSelection.collapsed(offset: next.length),
    );
    setState(() => _commandSuggestions = const []);
  }

  void _insertEmote(String code) {
    final ctrl = _chatController;
    final sel = ctrl.selection;
    final text = ctrl.text;
    final newText =
        text.substring(0, sel.baseOffset) +
        code +
        text.substring(sel.extentOffset);
    ctrl.value = TextEditingValue(
      text: newText,
      selection: TextSelection.collapsed(offset: sel.baseOffset + code.length),
    );
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);
    final currentUser = ref.watch(currentUserProvider);

    return Column(
      children: [
        Container(height: 1, color: colors.border),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          child: Row(
            children: [
              Icon(
                FIcons.messageSquare,
                size: 16,
                color: colors.mutedForeground,
              ),
              const SizedBox(width: 6),
              Text(
                l10n.usersChatTitle,
                style: typography.sm.copyWith(fontWeight: FontWeight.w600),
              ),
            ],
          ),
        ),
        Expanded(
          child: ListView.builder(
            controller: _scrollController,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            itemCount: _lines.length,
            itemBuilder: (context, index) {
              final line = _lines[index];
              if (line is SystemLine) {
                return SystemLineWidget(text: line.text);
              }
              return ChatBubble(message: (line as RemoteLine).data);
            },
          ),
        ),
        if (currentUser != null)
          _ChatInput(
            controller: _chatController,
            suggestions: _commandSuggestions,
            onPickSuggestion: _applySuggestion,
            onSend: _sendMessage,
            onInsertEmote: _insertEmote,
          )
        else
          Padding(
            padding: const EdgeInsets.all(12),
            child: Text(
              l10n.usersChatPlaceholderLogin,
              style: typography.sm.copyWith(color: colors.mutedForeground),
              textAlign: TextAlign.center,
            ),
          ),
      ],
    );
  }
}

class _ChatInput extends StatelessWidget {
  final TextEditingController controller;
  final List<ChatCommandSuggestion> suggestions;
  final ValueChanged<ChatCommandSuggestion> onPickSuggestion;
  final VoidCallback onSend;
  final ValueChanged<String> onInsertEmote;

  const _ChatInput({
    required this.controller,
    required this.suggestions,
    required this.onPickSuggestion,
    required this.onSend,
    required this.onInsertEmote,
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final l10n = AppLocalizations.of(context);

    return Container(
      padding: const EdgeInsets.fromLTRB(12, 8, 12, 12),
      decoration: BoxDecoration(
        border: Border(top: BorderSide(color: colors.border)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        mainAxisSize: MainAxisSize.min,
        children: [
          if (suggestions.isNotEmpty) ...[
            ChatCommandSuggestionsList(
              suggestions: suggestions,
              onPick: onPickSuggestion,
            ),
            const SizedBox(height: 8),
          ],
          Row(
            children: [
              FButton.icon(
                onPress: () => showEmotePicker(context, onInsertEmote),
                child: const Text('😊', style: TextStyle(fontSize: 18)),
              ),
              const SizedBox(width: 4),
              Expanded(
                child: FTextField(
                  control: FTextFieldControl.managed(controller: controller),
                  hint: l10n.usersChatPlaceholderTyping,
                ),
              ),
              const SizedBox(width: 8),
              FButton.icon(
                onPress: onSend,
                child: const Icon(FIcons.send),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
