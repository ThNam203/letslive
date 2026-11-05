/// The live chat of a channel (web `users/[userId]/chat.tsx`): a title, the
/// room's lines (newest at the bottom) and the message box. Shown under the
/// player and also while the channel is offline, as on web, because the
/// room outlives the stream.
///
/// While the history loads a spinner shows; if it fails an error with a retry
/// replaces the list (the message box stays usable). New lines arrive in
/// real time through `chat_providers.dart`. Suggestions from the message box
/// are drawn over the bottom of the list, limited to the list's height.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/widgets/feed_section.dart';
import 'package:letslive_mobile/features/watch/chat/chat_providers.dart';
import 'package:letslive_mobile/features/watch/chat/presentation/chat_input.dart';
import 'package:letslive_mobile/features/watch/chat/presentation/chat_line_view.dart';

/// The chat of the room [roomId] (the streamer's user id). Fills the height
/// it is given. With [compact] (the keyboard is open) the title row is left
/// out to leave room for the lines.
class ChatPanel extends ConsumerStatefulWidget {
  const ChatPanel({required this.roomId, this.compact = false, super.key});

  final String roomId;
  final bool compact;

  @override
  ConsumerState<ChatPanel> createState() => _ChatPanelState();
}

class _ChatPanelState extends ConsumerState<ChatPanel> {
  // Suggestions published by the message box.
  final _hints = ValueNotifier<Widget?>(null);

  @override
  void dispose() {
    _hints.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final roomId = widget.roomId;
    final chat = ref.watch(chatControllerProvider(roomId));
    final lines = chat.value;
    final Widget body;
    if (chat.hasError && !chat.isLoading) {
      body = SingleChildScrollView(
        child: FeedError(
          error: chat.error!,
          onRetry: () => ref.invalidate(chatControllerProvider(roomId)),
        ),
      );
    } else if (lines == null) {
      body = const Center(child: CircularProgressIndicator());
    } else {
      // Reversed so the list stays pinned to the newest line.
      body = ListView.builder(
        reverse: true,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        itemCount: lines.length,
        itemBuilder: (_, i) => ChatLineView(line: lines[lines.length - 1 - i]),
      );
    }

    return Column(
      children: [
        if (!widget.compact) ...[
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: Align(
              alignment: Alignment.centerLeft,
              child: Text(
                l10n.chat_title,
                style: Theme.of(context).textTheme.titleSmall,
              ),
            ),
          ),
          const Divider(height: 1),
        ],
        Expanded(
          child: LayoutBuilder(
            builder: (context, box) => Stack(
              fit: StackFit.expand,
              children: [
                body,
                ValueListenableBuilder<Widget?>(
                  valueListenable: _hints,
                  builder: (context, hints, _) => hints == null
                      ? const SizedBox.shrink()
                      : Align(
                          alignment: Alignment.bottomCenter,
                          child: ConstrainedBox(
                            constraints: BoxConstraints(
                              maxHeight: box.maxHeight.clamp(0, 220),
                            ),
                            child: hints,
                          ),
                        ),
                ),
              ],
            ),
          ),
        ),
        const Divider(height: 1),
        ChatInput(roomId: roomId, hints: _hints),
      ],
    );
  }
}
