/// The Messages tab: the user's conversations, newest activity first (web
/// `/messages`, the list half). Each row shows the other person or group, the
/// last message with its time, an online dot and the unread count.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/core/widgets/feed_section.dart';
import 'package:letslive_mobile/features/messages/data/dm_models.dart';
import 'package:letslive_mobile/features/messages/messages_providers.dart';
import 'package:letslive_mobile/features/messages/presentation/conversation_display.dart';

/// Unread counts above this show as "99+" (web).
const int _maxShownUnread = 99;

/// The Messages tab screen.
class ConversationsScreen extends ConsumerWidget {
  const ConversationsScreen({super.key});

  /// The retry of a failed load: the list and the unread counts.
  static void _reload(WidgetRef ref) {
    ref
      ..invalidate(dmUnreadProvider)
      ..invalidate(conversationsProvider);
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final conversations = ref.watch(conversationsProvider);
    final feed = conversations.value;
    final status = feedStatusBody(
      value: conversations.whenData((feed) => feed.items),
      onRetry: () => _reload(ref),
      empty: EmptyNotice(
        icon: Icons.chat_bubble_outline,
        title: l10n.messages_no_conversations_yet,
      ),
    );
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.messages_title),
        actions: [
          IconButton(
            key: const ValueKey('new-conversation'),
            tooltip: l10n.messages_new_conversation,
            icon: const Icon(Icons.add),
            onPressed: () => context.push(AppRoutes.newConversation),
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () {
          // The unread counts too: a failed load of them would otherwise
          // stay failed for the whole session.
          ref.invalidate(dmUnreadProvider);
          return ref.refresh(conversationsProvider.future);
        },
        child: status != null
            ? ListView(children: [status])
            : ListView.builder(
                itemCount: feed!.items.length + 1,
                itemBuilder: (context, index) => index == feed.items.length
                    ? PagedFooter<Conversation>(
                        feed: feed,
                        onLoadMore: () =>
                            ref.read(conversationsProvider.notifier).loadMore(),
                      )
                    : ConversationTile(
                        key: ValueKey('conversation-${feed.items[index].id}'),
                        conversation: feed.items[index],
                      ),
              ),
      ),
    );
  }
}

/// One row of the list. Text is clipped with an ellipsis so a long name or
/// message never overflows, even on a narrow screen with large text.
class ConversationTile extends ConsumerWidget {
  const ConversationTile({required this.conversation, super.key});

  final Conversation conversation;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final display = conversationDisplay(
      conversation,
      ref.watch(currentUserIdProvider),
      l10n,
    );
    final otherId = display.otherUserId;
    final isOnline =
        otherId != null &&
        ref.watch(dmPresenceProvider.select((p) => p.online.contains(otherId)));
    final unread =
        ref.watch(dmUnreadProvider.select((u) => u.value?[conversation.id])) ??
        0;
    final last = conversation.lastMessage;
    final muted = theme.colorScheme.onSurfaceVariant;
    return InkWell(
      onTap: () => context.push(AppRoutes.conversation(conversation.id)),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Row(
          children: [
            AvatarWithDot(
              pictureUrl: display.avatarUrl,
              name: display.name,
              isOnline: isOnline,
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          display.name,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: theme.textTheme.titleSmall,
                        ),
                      ),
                      if (last != null) ...[
                        const SizedBox(width: 8),
                        Flexible(
                          child: Text(
                            formatConversationTime(
                              last.createdAt,
                              DateTime.now(),
                              Localizations.localeOf(context).toString(),
                              l10n,
                            ),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: muted,
                            ),
                          ),
                        ),
                      ],
                    ],
                  ),
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          last?.text ?? l10n.messages_no_messages_yet,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: muted,
                          ),
                        ),
                      ),
                      if (unread > 0) ...[
                        const SizedBox(width: 8),
                        _UnreadPill(count: unread),
                      ],
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _UnreadPill extends StatelessWidget {
  const _UnreadPill({required this.count});

  final int count;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Container(
      constraints: const BoxConstraints(minWidth: 20, minHeight: 20),
      padding: const EdgeInsets.symmetric(horizontal: 6),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: scheme.primary,
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        count > _maxShownUnread ? '$_maxShownUnread+' : '$count',
        style: TextStyle(color: scheme.onPrimary, fontSize: 12),
      ),
    );
  }
}
