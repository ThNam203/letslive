/// One open conversation (web `/messages/[conversationId]`, the thread half):
/// a header with the person or group and their online state, the messages
/// (newest at the bottom, older ones load as you scroll up), who is typing
/// and the composer.
///
/// Long-pressing one of your own messages offers edit and delete. The screen
/// owns [dmThreadProvider] for its conversation, which is what marks the
/// conversation read while it is open.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/widgets/back_notice.dart';
import 'package:letslive_mobile/features/home/presentation/feed_section.dart';
import 'package:letslive_mobile/features/messages/data/dm_models.dart';
import 'package:letslive_mobile/features/messages/domain/dm_rules.dart';
import 'package:letslive_mobile/features/messages/messages_providers.dart';
import 'package:letslive_mobile/features/messages/presentation/conversation_display.dart';
import 'package:letslive_mobile/features/messages/presentation/message_bubble.dart';
import 'package:letslive_mobile/features/messages/presentation/message_composer.dart';
import 'package:letslive_mobile/features/profile/profile_providers.dart';
import 'package:letslive_mobile/features/vod/presentation/action_feedback.dart';

const List<String> _noTyping = [];

/// The conversation [conversationId].
class ConversationScreen extends ConsumerWidget {
  const ConversationScreen({required this.conversationId, super.key});

  /// From the route.
  final String conversationId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final conversation = ref.watch(conversationProvider(conversationId));
    final thread = ref.watch(dmThreadProvider(conversationId));
    final found = conversation.value;

    final Widget body;
    if (conversation.hasError && !conversation.isLoading) {
      body = Center(
        child: FeedError(
          error: conversation.error!,
          onRetry: () => ref.invalidate(conversationProvider(conversationId)),
        ),
      );
    } else if (conversation.hasValue && found == null) {
      // Deleted, never existed, or the user is not a member (Review Focus 5).
      body = BackNotice(
        icon: Icons.chat_bubble_outline,
        message: l10n.api_response_res_err_conversation_not_found,
      );
    } else if (thread.hasError && !thread.isLoading) {
      body = Center(
        child: FeedError(
          error: thread.error!,
          onRetry: () => ref.invalidate(dmThreadProvider(conversationId)),
        ),
      );
    } else if (!thread.hasValue || found == null) {
      body = const Center(child: CircularProgressIndicator());
    } else {
      body = Column(
        children: [
          Expanded(
            child: _MessageList(
              conversationId: conversationId,
              thread: thread.requireValue,
            ),
          ),
          _TypingIndicator(conversationId: conversationId),
          MessageComposer(conversationId: conversationId),
        ],
      );
    }
    return Scaffold(
      appBar: AppBar(
        title: found == null ? const Text('') : _Header(conversation: found),
      ),
      body: body,
    );
  }
}

/// Avatar, name and "Online" / "Offline" / member count.
class _Header extends ConsumerWidget {
  const _Header({required this.conversation});

  final Conversation conversation;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final display = conversationDisplay(
      conversation,
      ref.watch(currentUserIdProvider),
      l10n,
    );
    final otherId = display.otherUserId;
    final isOnline =
        otherId != null &&
        ref.watch(dmPresenceProvider.select((p) => p.online.contains(otherId)));
    final status = conversation.type == ConversationType.dm
        ? (isOnline ? l10n.messages_online : l10n.messages_offline)
        : l10n.messages_members_count(conversation.participants.length);
    return Row(
      children: [
        AvatarWithDot(
          pictureUrl: display.avatarUrl,
          name: display.name,
          isOnline: isOnline,
          radius: 18,
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                display.name,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: Theme.of(context).textTheme.titleSmall,
              ),
              Text(
                status,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ],
          ),
        ),
      ],
    );
  }
}

/// "NAME is typing..." (web `TypingIndicator`).
class _TypingIndicator extends ConsumerWidget {
  const _TypingIndicator({required this.conversationId});

  final String conversationId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final names = ref.watch(
      dmPresenceProvider.select((p) => p.typing[conversationId] ?? _noTyping),
    );
    if (names.isEmpty) return const SizedBox.shrink();
    final l10n = AppLocalizations.of(context);
    final text = switch (names.length) {
      1 => l10n.messages_typing_one(names[0]),
      2 => l10n.messages_typing_two(names[0], names[1]),
      _ => l10n.messages_typing_many(names[0], names.length - 1),
    };
    return Align(
      alignment: Alignment.centerLeft,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
        child: Text(
          text,
          style: Theme.of(context).textTheme.bodySmall
              ?.copyWith(color: Theme.of(context).colorScheme.onSurfaceVariant),
        ),
      ),
    );
  }
}

/// The messages, newest at the bottom. The list is reversed, so index 0 is
/// the newest message and the last item (at the top) is the loader of older
/// messages.
class _MessageList extends ConsumerStatefulWidget {
  const _MessageList({required this.conversationId, required this.thread});

  final String conversationId;
  final DmThread thread;

  @override
  ConsumerState<_MessageList> createState() => _MessageListState();
}

class _MessageListState extends ConsumerState<_MessageList> {
  final _scroll = ScrollController();

  @override
  void dispose() {
    _scroll.dispose();
    super.dispose();
  }

  /// Follows a new message to the bottom (web does too).
  void _followNewest(AsyncValue<DmThread>? before, AsyncValue<DmThread> now) {
    final old = before?.value?.messages;
    final next = now.value?.messages;
    if (old == null || next == null || old.isEmpty || next.isEmpty) return;
    if (old.last.id != next.last.id && _scroll.hasClients) {
      _scroll.animateTo(
        0,
        duration: const Duration(milliseconds: 200),
        curve: Curves.easeOut,
      );
    }
  }

  Future<void> _showActions(DmMessage message) async {
    final l10n = AppLocalizations.of(context);
    final myId = ref.read(currentUserIdProvider);
    final action = await showModalBottomSheet<_MessageAction>(
      context: context,
      showDragHandle: true,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (canEditMessage(message, myId))
              ListTile(
                key: const ValueKey('edit-message'),
                leading: const Icon(Icons.edit_outlined),
                title: Text(l10n.messages_edit_message),
                onTap: () => Navigator.of(context).pop(_MessageAction.edit),
              ),
            if (canDeleteMessage(message, myId))
              ListTile(
                key: const ValueKey('delete-message'),
                leading: const Icon(Icons.delete_outline),
                title: Text(l10n.messages_delete_message),
                onTap: () => Navigator.of(context).pop(_MessageAction.delete),
              ),
          ],
        ),
      ),
    );
    if (action == null || !mounted) return;
    final thread = ref.read(dmThreadProvider(widget.conversationId).notifier);
    switch (action) {
      case _MessageAction.edit:
        final text = await _askNewText(message.text);
        if (text == null || text == message.text || !mounted) return;
        _report(await thread.edit(message.id, text));
      case _MessageAction.delete:
        if (!await _confirmDelete() || !mounted) return;
        _report(await thread.delete(message.id));
    }
  }

  void _report(ApiResponse<Object?>? response) {
    if (response != null && !response.success) {
      showActionError(context, response);
    }
  }

  Future<String?> _askNewText(String current) => showDialog<String>(
    context: context,
    builder: (context) => _EditDialog(initial: current),
  );

  Future<bool> _confirmDelete() async {
    final l10n = AppLocalizations.of(context);
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(l10n.messages_delete_message),
        content: Text(l10n.messages_delete_message_confirm),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: Text(l10n.messages_cancel),
          ),
          FilledButton(
            key: const ValueKey('confirm-delete'),
            onPressed: () => Navigator.of(context).pop(true),
            child: Text(l10n.messages_delete),
          ),
        ],
      ),
    );
    return confirmed ?? false;
  }

  @override
  Widget build(BuildContext context) {
    ref.listen(dmThreadProvider(widget.conversationId), _followNewest);
    final l10n = AppLocalizations.of(context);
    final messages = widget.thread.messages;
    final myId = ref.watch(currentUserIdProvider);
    if (messages.isEmpty && !widget.thread.hasMore) {
      return Center(child: Text(l10n.messages_no_messages_yet));
    }
    return ListView.builder(
      controller: _scroll,
      reverse: true,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      itemCount: messages.length + 1,
      itemBuilder: (context, i) {
        if (i == messages.length) {
          return _OlderLoader(
            thread: widget.thread,
            // The loader asks from a microtask, which can run after the
            // screen is gone.
            onLoad: () {
              if (!mounted) return;
              ref
                  .read(dmThreadProvider(widget.conversationId).notifier)
                  .loadOlder();
            },
          );
        }
        final index = messages.length - 1 - i;
        final message = messages[index];
        final previous = index > 0 ? messages[index - 1] : null;
        final hasActions =
            canEditMessage(message, myId) || canDeleteMessage(message, myId);
        return MessageBubble(
          key: ValueKey(message.id),
          message: message,
          isOwn: message.senderId == myId,
          showSender: previous == null || previous.senderId != message.senderId,
          onLongPress: hasActions ? () => _showActions(message) : null,
          onOpenImage: (url) => ref.read(urlOpenerProvider)(Uri.parse(url)),
        );
      },
    );
  }
}

enum _MessageAction { edit, delete }

/// Top of the list: a spinner while older messages load (building it is the
/// "scrolled up to the top" signal, like `PagedFooter`), a retry after a
/// failure, "Beginning of conversation" once everything is loaded.
class _OlderLoader extends StatelessWidget {
  const _OlderLoader({required this.thread, required this.onLoad});

  final DmThread thread;
  final VoidCallback onLoad;

  @override
  Widget build(BuildContext context) {
    final error = thread.olderError;
    if (error != null) return FeedError(error: error, onRetry: onLoad);
    if (!thread.hasMore) {
      return Padding(
        padding: const EdgeInsets.all(8),
        child: Text(
          AppLocalizations.of(context).messages_beginning_of_conversation,
          textAlign: TextAlign.center,
          style: Theme.of(context).textTheme.bodySmall,
        ),
      );
    }
    // Not during build: loading changes provider state.
    Future.microtask(onLoad);
    return const Padding(
      padding: EdgeInsets.all(8),
      child: Center(child: CircularProgressIndicator()),
    );
  }
}

/// Asks for the new text of a message.
class _EditDialog extends StatefulWidget {
  const _EditDialog({required this.initial});

  final String initial;

  @override
  State<_EditDialog> createState() => _EditDialogState();
}

class _EditDialogState extends State<_EditDialog> {
  late final TextEditingController _text = TextEditingController(
    text: widget.initial,
  );

  @override
  void dispose() {
    _text.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return AlertDialog(
      title: Text(l10n.messages_edit_message),
      content: TextField(
        key: const ValueKey('edit-message-input'),
        controller: _text,
        autofocus: true,
        inputFormatters: [dmMessageLimitFormatter],
        maxLines: 5,
        minLines: 1,
        onChanged: (_) => setState(() {}),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: Text(l10n.messages_cancel),
        ),
        FilledButton(
          key: const ValueKey('save-message'),
          // Read when pressed, not when built, so it is never stale.
          onPressed: _text.text.trim().isEmpty
              ? null
              : () => Navigator.of(context).pop(_text.text.trim()),
          child: Text(l10n.messages_save),
        ),
      ],
    );
  }
}
